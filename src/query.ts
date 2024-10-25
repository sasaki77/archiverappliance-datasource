import _ from 'lodash';
import { MutableDataFrame, getFieldDisplayName } from '@grafana/data';

import { applyFunctionDefs } from './aafunc';
import { TargetQuery } from './types';
import { AAclient } from 'aaclient';
import { responseParse } from 'responseParse';

export function doQuery(aaclient: AAclient, targets: TargetQuery[]): Promise<{ data: Array<MutableDataFrame<any>> }> {
  // Create promises to buil URLs for each targets: [[URLs for target 1], [URLs for target 2] , ...]
  const urlsArray = _.map(targets, (target) => aaclient.buildUrls(target));

  // Wait for building URLs then create target data
  const targetProcesses = Promise.all(urlsArray).then((urlsArray) => {
    // Create promises to retrieve data for each targets: [[Responses for target 1], [Reponses for target 2] , ...]
    const responsePromisesArray = aaclient.createUrlRequests(urlsArray);

    // Data processing for each targets: [[Processed data for target 1], [Processed data for target 2], ...]
    const targetProcesses = _.map(responsePromisesArray, (responsePromises, i) => {
      return Promise.all(responsePromises).then((responses) => targetProcess(responses, targets[i]));
    });

    // Wait all target data processings
    return Promise.all(targetProcesses);
  });

  return targetProcesses.then((dataFramesArray) => postProcess(dataFramesArray));
}

export async function setAlias(dataFrames: MutableDataFrame[], target: TargetQuery): Promise<MutableDataFrame[]> {
  if (!target.alias) {
    return Promise.resolve(dataFrames);
  }

  let pattern: RegExp;
  if (target.aliasPattern) {
    pattern = new RegExp(target.aliasPattern, '');
  }

  const newDataFrames = _.map(dataFrames, (dataFrame) => {
    const valfields = _.filter(dataFrame.fields, (field) => field.name !== 'time' && field.name !== 'index');

    const newValfields = _.map(valfields, (valfield) => {
      const displayName = getFieldDisplayName(valfield, dataFrame);
      const alias = pattern ? displayName.replace(pattern, target.alias) : target.alias;

      return {
        ...valfield,
        config: {
          ...valfield.config,
          displayName: alias,
        },
        state: {
          ...valfield.state,
          displayName: alias,
        },
      };
    });

    return new MutableDataFrame({
      ...dataFrame,
      fields: [dataFrame.fields[0]].concat(newValfields),
    });
  });

  return Promise.resolve(newDataFrames);
}

export function applyFunctions(dataFrames: MutableDataFrame[], target: TargetQuery) {
  if (target.functions === undefined) {
    return Promise.resolve(dataFrames);
  }

  return applyFunctionDefs(target.functions, dataFrames);
}

function targetProcess(responses: any, target: TargetQuery) {
  return responseParse(responses, target)
    .then((dataFrames) => setAlias(dataFrames, target))
    .then((dataFrames) => applyFunctions(dataFrames, target));
}

function postProcess(dataFramesArray: MutableDataFrame[][]) {
  const dataFrames = _.flatten(dataFramesArray);

  return { data: dataFrames };
}
