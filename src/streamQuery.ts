import _ from 'lodash';
import { Observable, Subscriber } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import ms, { StringValue } from 'ms';
import { DataQueryResponse, LoadingState, DataFrame } from '@grafana/data';

import { TargetQuery } from './types';
import { AAclient } from 'aaclient';
import { responseParse } from 'responseParse';
import { applyFunctions, setAlias } from 'query';

export const STREAM_FROM_MARGIN_MS = 2000;
export const STREAM_TO_MARGIN_MS = 500;

type StreamBuffer = {
  fields: { [key: string]: any[] };
  capacity: number;
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
      //   "PV:NAME": {
      //     fields: {
      //       time:  [t1, t2, t3, ...],
      //       value: [v1, v2, v3, ...],
      //       ... (other fields)
      //     }
      //     capacity: 1000
      //   }
      // }
      const buffers: { [key: string]: StreamBuffer } = {};

      doQueryStream(this.aaclient, targets, buffers)
        .then((data) => {
          subscriber.next(data);

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
    if (id in this.timerIDs) {
      clearTimeout(this.timerIDs[id]);
    }
    this.timerIDs = {};
  }
}

function doQueryStream(
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

function mergeToBuffers(
  dataFrames: DataFrame[],
  buffers: Record<string, StreamBuffer>,
  target: TargetQuery
): Promise<DataFrame[]> {
  const toTimestamp = target.to.getTime();

  const resultFrames = dataFrames
    .filter((f) => f.name !== undefined)
    .map((frame) => {
      const name = frame.name!;
      let buffer = buffers[name];

      // --- Initialize buffer (if first time) ---
      if (!buffer) {
        const defaultCap = Math.max(target.maxDataPoints, frame.length);
        const capacity = parseInt(target.strmCap, 10) || defaultCap;

        buffer = {
          fields: {},
          capacity,
        };

        for (const field of frame.fields) {
          buffer.fields[field.name] = [...field.values];
        }

        buffers[name] = buffer;

        // --- Build immutable DataFrame ---
        return buildDataFrame(frame, buffer);
      }

      const timeArray = buffer.fields['time'];

      // --- Append new data (diff update) ---
      for (let i = 0; i < frame.length; i++) {
        // Extract row
        const row: Record<string, any> = {};
        for (const field of frame.fields) {
          row[field.name] = field.values[i];
        }

        // Skip future data
        if (row.time > toTimestamp) {
          continue;
        }

        // Skip duplicate or old data
        if (timeArray && timeArray.length > 0) {
          const lastTime = timeArray[timeArray.length - 1];
          if (row.time <= lastTime) {
            continue;
          }
        }

        // Append row to buffer
        for (const field of frame.fields) {
          buffer.fields[field.name].push(row[field.name]);
        }
      }

      // --- Trim buffer (capacity control) ---
      for (const [fieldName, values] of Object.entries(buffer.fields)) {
        if (values.length > buffer.capacity) {
          buffer.fields[fieldName] = values.slice(values.length - buffer.capacity);
        }
      }

      // --- Build immutable DataFrame ---
      return buildDataFrame(frame, buffer);
    });

  return Promise.resolve(resultFrames);
}

function buildDataFrame(frame: DataFrame, buffer: StreamBuffer): DataFrame {
  return {
    ...frame,
    fields: frame.fields.map((field) => ({
      ...field,
      values: [...(buffer.fields[field.name] || [])],
    })),
    length: buffer.fields['time'].length,
  };
}
