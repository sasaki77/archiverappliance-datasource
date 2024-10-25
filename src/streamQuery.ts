import _ from 'lodash';
import { Observable, Subscriber } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import ms from 'ms';
import { CircularDataFrame, DataQueryResponse, LoadingState, MutableDataFrame } from '@grafana/data';

import { TargetQuery } from './types';
import { AAclient } from 'aaclient';
import { responseParse } from 'responseParse';
import { applyFunctions, setAlias } from 'query';

export class StreamQuery {
  aaclient: AAclient;
  timerIDs: { [key: string]: any };

  constructor(aaclient: AAclient) {
    this.aaclient = aaclient;
    this.timerIDs = {};
  }

  runStream(targets: TargetQuery[], streamTargets: TargetQuery[], intervalMs: number): Observable<DataQueryResponse> {
    return new Observable<DataQueryResponse>((subscriber) => {
      // Create new targets to disable auto Extrapolation
      const t = _.map(targets, (target) => {
        return {
          ...target,
          options: {
            ...target.options,
            disableExtrapol: 'true',
          },
        };
      });

      const id = uuidv4();
      const cirFrames: { [key: string]: CircularDataFrame<any> } = {};

      doQueryStream(this.aaclient, t, cirFrames).then((data) => {
        subscriber.next(data);

        const interval = (streamTargets[0].strmInt && ms(streamTargets[0].strmInt)) || intervalMs;

        // Create new targets to update interval time
        const new_t = _.map(t, (target) => {
          const t_int = target.interval ? Math.floor(interval / 1000).toFixed() : '';
          const int = interval >= 1000 ? t_int : '';

          return {
            ...target,
            interval: int,
          };
        });

        this.timerIDs[id] = setTimeout(this.timerLoop, interval, subscriber, new_t, id, cirFrames, interval);
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
    frames: { [key: string]: CircularDataFrame },
    interval: number
  ) => {
    updateTargetDate(targets);
    const data = await doQueryStream(this.aaclient, targets, frames);

    subscriber.next(data);
    if (id in this.timerIDs) {
      this.timerIDs[id] = setTimeout(this.timerLoop, interval, subscriber, targets, id, frames, interval);
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
  frames: { [key: string]: CircularDataFrame }
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
        .then((responses) => responseParse(responses, targets[i]))
        .then((dataFrames) => mergeResToCirFrames(dataFrames, frames, targets[i]))
        .then((dataFrames) => setAlias(dataFrames, targets[i]))
        .then((dataFrames) => applyFunctions(dataFrames, targets[i]));
    });

    // Wait all target data processings
    return Promise.all(targetProcesses);
  });

  return targetProcesses.then((dataFramesArray) => streamPostProcess(dataFramesArray));
}

function streamPostProcess(dataFramesArray: MutableDataFrame[][]) {
  const dataFrames = _.flatten(dataFramesArray);

  return { data: dataFrames, state: LoadingState.Streaming };
}

function updateTargetDate(targets: TargetQuery[]) {
  return _.map(targets, (target) => {
    // AA should probably not able to return latest data near the "now".
    // So, the time range is set from 2 secs ago from last update date and to 500 msecs ago from "now".
    target.from = new Date(target.to.getTime() - 2000);
    target.to = new Date(Date.now() - 500);
    return target;
  });
}

function mergeResToCirFrames(
  dataFrames: MutableDataFrame[],
  cirFrames: { [key: string]: CircularDataFrame },
  target: TargetQuery
): Promise<MutableDataFrame[]> {
  const to = target.to.getTime();
  const d = _.filter(dataFrames, (frame) => frame.name !== undefined);

  const frames = _.map(d, (frame) => {
    if (frame.name === undefined) {
      return frame;
    }

    // Create frame for new arrival data
    if (!(frame.name in cirFrames)) {
      cirFrames[frame.name] = createStreamFrame(target, frame);
      return cirFrames[frame.name];
    }

    const last_time = cirFrames[frame.name].get(cirFrames[frame.name].length - 1)['time'];

    // Update frame data
    for (let i = 0; i < frame.length; i++) {
      const fields = frame.get(i);
      if (fields['time'] <= last_time || fields['time'] > to) {
        continue;
      }
      cirFrames[frame.name].add(fields);
    }
    return cirFrames[frame.name];
  });

  return Promise.resolve(frames);
}

function createStreamFrame(target: TargetQuery, dataFrame: MutableDataFrame) {
  const c = parseInt(target.strmCap, 10);
  const cap = dataFrame.refId ? c || dataFrame.length : dataFrame.length;

  const new_frame = new CircularDataFrame({
    append: 'tail',
    capacity: cap,
  });

  new_frame.name = dataFrame.name;
  for (const field of dataFrame.fields) {
    new_frame.addField(field);
  }

  return new_frame;
}
