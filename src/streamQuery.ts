import _ from 'lodash';
import { Observable, Subscriber } from 'rxjs';
import ms, { StringValue } from 'ms';
import { DataQueryResponse, LoadingState, DataFrame } from '@grafana/data';

import { TargetQuery } from './types';
import { AAclient } from 'aaclient';
import { isExtrapolated, responseParse } from 'responseParse';
import { applyFunctions, setAlias } from 'query';

export const STREAM_FROM_MARGIN_MS = 2000;
export const STREAM_TO_MARGIN_MS = 500;

export class StreamQuery {
  aaclient: AAclient;
  timerIDs: { [key: string]: any };
  private nextStreamID = 0;

  constructor(aaclient: AAclient) {
    this.aaclient = aaclient;
    this.timerIDs = {};
  }

  runStream(targets: TargetQuery[], streamTargets: TargetQuery[], intervalMs: number): Observable<DataQueryResponse> {
    return new Observable<DataQueryResponse>((subscriber) => {
      const id = String(this.nextStreamID++);

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
        buffer = new StreamBuffer(frame, capacity, target.refId);
        buffers[key] = buffer;

        // --- Build immutable DataFrame ---
        return buffer.toDataFrame(frame);
      }

      // A target new to the buffer is shown everything it holds, rather than
      // only the last points of its own capacity.
      const firstMerge = buffer.addTarget(target.refId, capacity);

      // --- Append new data (diff update) ---
      const times = frame.fields.find((field) => field.name === 'time')?.values;
      let lastTime = buffer.newestTime();

      for (let i = 0; i < frame.length; i++) {
        const time = times?.[i];

        // Skip future data, and data already held
        if (time !== undefined && (time > toTimestamp || (lastTime !== undefined && time <= lastTime))) {
          continue;
        }

        buffer.append(frame, i);
        lastTime = time;
      }

      // --- Build immutable DataFrame ---
      return buffer.toDataFrame(frame, firstMerge ? undefined : buffer.capacityOf(target.refId));
    });

  return Promise.resolve(resultFrames);
}

// The points of one time series, held in a ring: its arrays are allocated at
// the capacity and written by index, so a tick neither grows them nor copies
// them. Only the frame handed out is copied.
class StreamBuffer {
  // One capacity per target sharing the buffer, since strmCap is set per target:
  // the ring is as long as the largest, and a target is shown at most its own.
  private capacities: { [refId: string]: number };
  private fields: { [name: string]: any[] } = {};
  private size: number;
  private start = 0;
  private length: number;

  // The ring holds the initial query in full, even when it is longer than the
  // capacity, which the first tick then cuts it back to.
  constructor(frame: DataFrame, capacity: number, refId: string) {
    this.capacities = { [refId]: capacity };
    this.size = Math.max(capacity, frame.length);
    this.length = frame.length;

    for (const field of frame.fields) {
      const values = new Array(this.size).fill(0);
      for (let i = 0; i < frame.length; i++) {
        values[i] = field.values[i];
      }
      this.fields[field.name] = values;
    }
  }

  // Registers the capacity this target asks for and cuts the ring back to the
  // largest capacity the targets sharing it ask for.
  // Returns whether the target is new to the buffer.
  addTarget(refId: string, capacity: number): boolean {
    const firstMerge = !(refId in this.capacities);
    if (firstMerge) {
      this.capacities[refId] = capacity;
    }

    const size = Math.max(...Object.values(this.capacities));
    if (this.size !== size) {
      this.resize(size);
    }

    return firstMerge;
  }

  // How many of the newest points this target is shown.
  capacityOf(refId: string): number {
    return this.capacities[refId];
  }

  append(frame: DataFrame, row: number) {
    const index = (this.start + this.length) % this.size;

    for (const field of frame.fields) {
      this.fields[field.name][index] = field.values[row];
    }

    if (this.length < this.size) {
      this.length += 1;
    } else {
      // The row just written took the place of the oldest one.
      this.start = (this.start + 1) % this.size;
    }
  }

  // The time of the newest point held, or undefined when the buffer is empty or
  // holds no time field, as the index array format does not.
  newestTime() {
    const times = this.fields['time'];
    if (!times || this.length === 0) {
      return undefined;
    }
    return times[(this.start + this.length - 1) % this.size];
  }

  // Without a count the whole buffer is returned, so that a target's initial
  // query is shown in full even when it holds more points than the capacity.
  toDataFrame(frame: DataFrame, count = Infinity): DataFrame {
    const length = Math.min(count, this.length);
    const from = (this.start + this.length - length) % this.size;

    return {
      ...frame,
      fields: frame.fields.map((field) => ({
        ...field,
        values: this.read(field.name, from, length),
      })),
      length,
    };
  }

  private resize(size: number) {
    const length = Math.min(this.length, size);
    const dropped = this.length - length;

    for (const [name, values] of Object.entries(this.fields)) {
      const resized = new Array(size).fill(0);
      for (let i = 0; i < length; i++) {
        resized[i] = values[(this.start + dropped + i) % this.size];
      }
      this.fields[name] = resized;
    }

    this.size = size;
    this.start = 0;
    this.length = length;
  }

  private read(name: string, from: number, length: number) {
    const values = this.fields[name] ?? [];
    const end = from + length;

    if (end <= this.size) {
      return values.slice(from, end);
    }
    return values.slice(from, this.size).concat(values.slice(0, end - this.size));
  }
}
