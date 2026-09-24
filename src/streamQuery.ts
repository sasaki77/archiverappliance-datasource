import _ from 'lodash';
import { Observable, Subscriber } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import ms, { StringValue } from 'ms';
import { DataQueryResponse, LoadingState, DataFrame } from '@grafana/data';

import { TargetQuery } from './types';
import { AAclient } from 'aaclient';
import { isExtrapolated, responseParse } from 'responseParse';
import { applyFunctions, setAlias } from 'query';

export const STREAM_FROM_MARGIN_MS = 2000;
export const STREAM_TO_MARGIN_MS = 500;

// A ring: its arrays are allocated at the capacity and written by index, so a
// tick neither grows them nor copies them. Only the frame handed out is copied.
type StreamBuffer = {
  fields: { [key: string]: any[] };
  // One capacity per target sharing the buffer, since strmCap is set per target:
  // the ring is as long as the largest, and a target is shown at most its own.
  capacities: { [refId: string]: number };
  size: number;
  start: number;
  length: number;
};

export class StreamQuery {
  aaclient: AAclient;
  timerIDs: { [key: string]: any };

  constructor(aaclient: AAclient) {
    this.aaclient = aaclient;
    this.timerIDs = {};
  }

  runStream(targets: TargetQuery[], streamTargets: TargetQuery[], intervalMs: number): Observable<DataQueryResponse> {
    return new Observable<DataQueryResponse>((subscriber) => {
      const id = uuidv4();

      // Buffer structure per time series (mutable internal state)
      //
      // buffers = {
      //   "PV:NAME\0mean\0...": {
      //     fields: {
      //       time:  [t1, t2, t3, ...],
      //       value: [v1, v2, v3, ...],
      //       ... (other fields)
      //     }
      //     capacities: { A: 1000, B: 500 }
      //   }
      // }
      const buffers: { [key: string]: StreamBuffer } = {};

      doQueryStream(this.aaclient, targets, buffers)
        .then((data) => {
          subscriber.next(data);

          // next() may have run the teardown, before there was a timer to clear.
          if (subscriber.closed) {
            return;
          }

          const interval = (streamTargets[0].strmInt && ms(streamTargets[0].strmInt as StringValue)) || intervalMs;

          const newTargets = _.map(targets, (target) => {
            const t_int = target.interval ? Math.floor(interval / 1000).toFixed() : '';
            const int = interval >= 1000 ? t_int : '';

            return {
              ...target,
              interval: int,
            };
          });

          this.timerIDs[id] = setTimeout(this.timerLoop, interval, subscriber, newTargets, id, buffers, interval);
        })
        .catch((err) => {
          subscriber.error({
            message: 'Failed to fetch streaming data',
            status: 'error',
            statusText: err.data?.message || err.message || 'Unknown error',
          });
        });

      return () => {
        this.timerClear(id);
      };
    });
  }

  private timerLoop = async (
    subscriber: Subscriber<DataQueryResponse>,
    targets: TargetQuery[],
    id: string,
    buffers: { [key: string]: StreamBuffer },
    interval: number
  ) => {
    const updatedTargets = updateTargetDate(targets);

    try {
      const data = await doQueryStream(this.aaclient, updatedTargets, buffers);

      subscriber.next(data);

      if (id in this.timerIDs) {
        this.timerIDs[id] = setTimeout(this.timerLoop, interval, subscriber, updatedTargets, id, buffers, interval);
      }
    } catch (err) {
      subscriber.error({
        message: 'Failed to fetch streaming data',
        status: 'error',
        statusText: 'Failed to fetch streaming data',
      });
    }
  };

  private timerClear(id: string) {
    clearTimeout(this.timerIDs[id]);
    delete this.timerIDs[id];
  }
}

export function doQueryStream(
  aaclient: AAclient,
  targets: TargetQuery[],
  buffers: { [key: string]: StreamBuffer }
): Promise<DataQueryResponse> {
  // Create promises to buil URLs for each targets: [[URLs for target 1], [URLs for target 2] , ...]
  const urlsArray = _.map(targets, (target) => aaclient.buildUrls(target));

  // Wait for building URLs then create target data
  const targetProcesses = Promise.all(urlsArray).then((urlsArray) => {
    // Create promises to retrieve data for each targets: [[Responses for target 1], [Reponses for target 2] , ...]
    const responsePromisesArray = aaclient.createUrlRequests(urlsArray);

    // Data processing for each targets: [[Processed data for target 1], [Processed data for target 2], ...]
    const targetProcesses = _.map(responsePromisesArray, (responsePromises, i) => {
      return Promise.all(responsePromises)
        .then((responses) => responseParse(responses, targets[i], true))
        .then((dataFrames) => mergeToBuffers(dataFrames, buffers, targets[i]))
        .then((dataFrames) => extrapolate(dataFrames, targets[i]))
        .then((dataFrames) => setAlias(dataFrames, targets[i]))
        .then((dataFrames) => applyFunctions(dataFrames, targets[i]));
    });

    // Wait all target data processings
    return Promise.all(targetProcesses);
  });

  return targetProcesses.then((dataFramesArray) => streamPostProcess(dataFramesArray));
}

function streamPostProcess(dataFramesArray: DataFrame[][]) {
  const dataFrames = _.flatten(dataFramesArray);
  return { data: dataFrames, state: LoadingState.Streaming };
}

function updateTargetDate(targets: TargetQuery[]) {
  return _.map(targets, (target) => ({
    // AA should probably not able to return latest data near the "now".
    // So, the time range is set from 2 secs ago from last update date and to 500 msecs ago from "now".
    ...target,
    from: new Date(target.to.getTime() - STREAM_FROM_MARGIN_MS),
    to: new Date(Date.now() - STREAM_TO_MARGIN_MS),
  }));
}

// Only the frame handed out carries the extrapolated point: in the buffer it
// would pile up one fake sample per tick while the PV does not change. It stops
// short of the last STREAM_FROM_MARGIN_MS, which the next query fetches again
// because the archiver may not have those samples yet.
function extrapolate(dataFrames: DataFrame[], target: TargetQuery): DataFrame[] {
  if (!isExtrapolated(target)) {
    return dataFrames;
  }

  const extrapolatedTime = target.to.getTime() - STREAM_FROM_MARGIN_MS - 1;

  for (const frame of dataFrames) {
    const last = frame.length - 1;
    if (frame.fields[0]?.name !== 'time' || last < 0 || frame.fields[0].values[last] > extrapolatedTime) {
      continue;
    }

    // buildDataFrame gave this frame its own copy of the values.
    frame.fields.forEach((field, i) => field.values.push(i === 0 ? extrapolatedTime : field.values[last]));
    frame.length += 1;
  }

  return dataFrames;
}

// Targets that fetch the same series share a buffer, since what differs between
// them (alias, functions) is applied after the merge. The auto interval is left
// out: it changes after the first query, and it is the same for every target.
function bufferKey(frame: DataFrame, target: TargetQuery) {
  return [
    frame.name,
    target.operator,
    target.options.binInterval ?? '',
    target.options.disableAutoRaw ?? '',
    target.options.arrayFormat ?? '',
    frame.fields[1]?.config?.displayName ?? '',
  ].join('\u0000');
}

function mergeToBuffers(
  dataFrames: DataFrame[],
  buffers: Record<string, StreamBuffer>,
  target: TargetQuery
): Promise<DataFrame[]> {
  const toTimestamp = target.to.getTime();

  const resultFrames = dataFrames
    .filter((f) => f.name !== undefined)
    .map((frame) => {
      const key = bufferKey(frame, target);
      const capacity = parseInt(target.strmCap, 10) || Math.max(target.maxDataPoints, frame.length);
      let buffer = buffers[key];

      // --- Initialize buffer (if first time) ---
      if (!buffer) {
        buffer = createBuffer(frame, capacity, target.refId);
        buffers[key] = buffer;

        // --- Build immutable DataFrame ---
        return buildDataFrame(frame, buffer);
      }

      // A target absent from capacities has not been shown this buffer before,
      // so it is given everything the buffer holds rather than its last points.
      const firstMerge = !(target.refId in buffer.capacities);
      if (firstMerge) {
        buffer.capacities[target.refId] = capacity;
      }

      // --- Resize the ring (capacity control) ---
      const bufferCapacity = Math.max(...Object.values(buffer.capacities));
      if (buffer.size !== bufferCapacity) {
        resizeBuffer(buffer, bufferCapacity);
      }

      // --- Append new data (diff update) ---
      const times = frame.fields.find((field) => field.name === 'time')?.values;
      let lastTime = newestTime(buffer);

      for (let i = 0; i < frame.length; i++) {
        const time = times?.[i];

        // Skip future data, and data already held
        if (time !== undefined && (time > toTimestamp || (lastTime !== undefined && time <= lastTime))) {
          continue;
        }

        appendRow(buffer, frame, i);
        lastTime = time;
      }

      // --- Build immutable DataFrame ---
      return buildDataFrame(frame, buffer, firstMerge ? bufferCapacity : buffer.capacities[target.refId]);
    });

  return Promise.resolve(resultFrames);
}

// The ring holds the initial query in full, even when it is longer than the
// capacity, which the first tick then cuts it back to.
function createBuffer(frame: DataFrame, capacity: number, refId: string): StreamBuffer {
  const size = Math.max(capacity, frame.length);
  const fields: { [key: string]: any[] } = {};

  for (const field of frame.fields) {
    const values = new Array(size).fill(0);
    for (let i = 0; i < frame.length; i++) {
      values[i] = field.values[i];
    }
    fields[field.name] = values;
  }

  return { fields, capacities: { [refId]: capacity }, size, start: 0, length: frame.length };
}

function resizeBuffer(buffer: StreamBuffer, size: number) {
  const length = Math.min(buffer.length, size);
  const dropped = buffer.length - length;

  for (const [name, values] of Object.entries(buffer.fields)) {
    const resized = new Array(size).fill(0);
    for (let i = 0; i < length; i++) {
      resized[i] = values[(buffer.start + dropped + i) % buffer.size];
    }
    buffer.fields[name] = resized;
  }

  buffer.size = size;
  buffer.start = 0;
  buffer.length = length;
}

function appendRow(buffer: StreamBuffer, frame: DataFrame, row: number) {
  const index = (buffer.start + buffer.length) % buffer.size;

  for (const field of frame.fields) {
    buffer.fields[field.name][index] = field.values[row];
  }

  if (buffer.length < buffer.size) {
    buffer.length += 1;
  } else {
    // The row just written took the place of the oldest one.
    buffer.start = (buffer.start + 1) % buffer.size;
  }
}

// The time of the newest point held, or undefined when the buffer is empty or
// holds no time field, as the index array format does not.
function newestTime(buffer: StreamBuffer) {
  const times = buffer.fields['time'];
  if (!times || buffer.length === 0) {
    return undefined;
  }
  return times[(buffer.start + buffer.length - 1) % buffer.size];
}

// Without a count the whole buffer is returned, so that a target's initial
// query is shown in full even when it holds more points than the capacity.
function buildDataFrame(frame: DataFrame, buffer: StreamBuffer, count = buffer.length): DataFrame {
  const length = Math.min(count, buffer.length);
  const from = (buffer.start + buffer.length - length) % buffer.size;

  return {
    ...frame,
    fields: frame.fields.map((field) => ({
      ...field,
      values: readRange(buffer, field.name, from, length),
    })),
    length,
  };
}

function readRange(buffer: StreamBuffer, name: string, from: number, length: number) {
  const values = buffer.fields[name] ?? [];
  const end = from + length;

  if (end <= buffer.size) {
    return values.slice(from, end);
  }
  return values.slice(from, buffer.size).concat(values.slice(0, end - buffer.size));
}
