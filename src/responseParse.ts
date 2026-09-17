import _ from 'lodash';
import { createDataFrame, DataFrame, FieldType, addRow } from '@grafana/data';

import { getToScalarFuncs } from './aafunc';
import { TargetQuery, AADataQueryData, AADataQueryResponse, isNumberArray } from './types';

export function isExtrapolated(target: TargetQuery) {
  return (target.operator === 'raw' || target.interval === '') && target.options.disableExtrapol !== 'true';
}

// A stream extrapolates after merging into its buffer, not here.
export function responseParse(responses: AADataQueryResponse[], target: TargetQuery, stream = false) {
  const dataFramesArray = _.map(responses, (response) => {
    const dataFrames = _.map(response.data, (targetRes) => {
      if (targetRes.meta.waveform) {
        const toScalarFuncs = getToScalarFuncs(target.functions);
        if (toScalarFuncs.length > 0) {
          return parseArrayResponseToScalar(targetRes, toScalarFuncs, target);
        }
        return parseArrayResponse(targetRes, target);
      }
      return parseScalarResponse(targetRes, target);
    });

    return _.flatten(dataFrames);
  });

  const dataFrames = _.flatten(dataFramesArray);

  if (stream || !isExtrapolated(target)) {
    return Promise.resolve(dataFrames);
  }

  // Extrapolation for raw operator
  const to_msec = target.to.getTime();
  const extrapolationDataFrames = _.map(dataFrames, (dataframe) => {
    if (dataframe.fields[0].name !== 'time') {
      return dataframe;
    }

    //const latestval = dataframe.get(dataframe.length - 1);
    const newRow = [];
    const lastindex = dataframe.length - 1;

    for (const field of dataframe.fields) {
      newRow.push(field.values[lastindex]);
    }

    // first field of newRow is time field
    newRow[0] = to_msec;

    addRow(dataframe, newRow);

    return dataframe;
  });

  return Promise.resolve(extrapolationDataFrames);
}

function parseArrayResponseToScalar(
  targetRes: AADataQueryData,
  toScalarFuncs: Array<{ func: any; label: string }>,
  target: TargetQuery
) {
  // Type check for columnValues
  if (!isNumberArray(targetRes)) {
    return [];
  }

  const frames = _.map(toScalarFuncs, (func) => {
    const values = _.map(targetRes.data, (datapoint) => func.func(datapoint.val));
    const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
    const frame = createDataFrame({
      refId: target.refId,
      name: targetRes.meta.name,
      fields: [
        { name: 'time', type: FieldType.time, values: times },
        {
          name: 'value',
          type: FieldType.number,
          values: values,
          config: { displayName: `${targetRes.meta.name}(${func.label})` },
        },
      ],
    });
    return frame;
  });

  return frames;
}

function parseArrayResponse(targetRes: AADataQueryData, target: TargetQuery) {
  let fields;
  if (target.options.arrayFormat === 'dt-space') {
    fields = makeDtSpaceArrayFields(targetRes);
  } else if (target.options.arrayFormat === 'index') {
    fields = makeIndexArrayFields(targetRes);
  } else {
    fields = makeTimeseriesArrayFields(targetRes);
  }

  // A waveform that is not numeric has no fields to show. A frame without them
  // would break extrapolation and setAlias, which both read the first field.
  if (fields.length === 0) {
    return [];
  }

  const frame = createDataFrame({
    refId: target.refId,
    name: targetRes.meta.name,
    fields,
  });

  return frame;
}

function makeDtSpaceArrayFields(targetRes: AADataQueryData) {
  // Type check for columnValues
  if (!isNumberArray(targetRes)) {
    return [];
  }

  const targetData = targetRes.data;

  const field_val = _.reduce(
    targetData,
    (fields, data, i) => {
      fields['vals'] = fields['vals'].concat(data.val);

      const len = data.val.length;
      for (let i = 0; i < len; i++) {
        const date = data.millis + i;
        fields['times'].push(date);
      }

      return fields;
    },
    { times: [], vals: [] } as { times: number[]; vals: number[] }
  );

  const fields = [
    { name: 'time', type: FieldType.time, values: field_val['times'] },
    { name: targetRes.meta.name, type: FieldType.number, values: field_val['vals'] },
  ];

  return fields;
}

// A waveform changes length between samples, as its NORD does. Every sample is
// laid out to the longest one, with null where it has no element, so that the
// frame keeps every element that was archived.
function waveformWidth(samples: Array<{ val: number[] }>) {
  return samples.reduce((width, sample) => Math.max(width, sample.val.length), 0);
}

function padWaveform(val: number[], width: number): Array<number | null> {
  return Array.from({ length: width }, (_v, i) => (i < val.length ? val[i] : null));
}

function makeIndexArrayFields(targetRes: AADataQueryData) {
  // Type check for columnValues
  if (!isNumberArray(targetRes)) {
    return [];
  }

  const targetData = targetRes.data;
  const width = waveformWidth(targetData);

  const fields: Array<{ name: string; type: FieldType; values: Array<number | null> }> = [
    { name: 'index', type: FieldType.number, values: Array.from({ length: width }, (_v, i) => i) },
  ];

  for (const data of targetData) {
    fields.push({
      name: toISOStringWithTimezone(new Date(data.millis)),
      type: FieldType.number,
      values: padWaveform(data.val, width),
    });
  }

  return fields;
}

function makeTimeseriesArrayFields(targetRes: AADataQueryData) {
  // Type check for columnValues
  if (!isNumberArray(targetRes)) {
    return [];
  }

  const targetData = targetRes.data;
  const width = waveformWidth(targetData);

  const times: number[] = targetData.map((datapoint) => datapoint.millis);
  const fields: Array<{ name: string; type: FieldType; values: Array<number | null> }> = [
    { name: 'time', type: FieldType.time, values: times },
  ];

  // Add fields for each waveform elements
  for (let i = 0; i < width; i++) {
    fields.push({
      name: `${targetRes.meta.name}[${i}]`,
      type: FieldType.number,
      values: targetData.map((datapoint) => (i < datapoint.val.length ? datapoint.val[i] : null)),
    });
  }

  return fields;
}

function parseScalarResponse(targetRes: AADataQueryData, target: TargetQuery): DataFrame {
  const values = _.map(targetRes.data, (datapoint) => datapoint.val);
  const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
  const frame = createDataFrame({
    refId: target.refId,
    name: targetRes.meta.name,
    fields: [
      { name: 'time', type: FieldType.time, values: times },
      { name: 'value', type: FieldType.number, values: values, config: { displayName: targetRes.meta.name } },
    ],
  });
  return frame;
}

// ISO 8601 format Date with Timezone Offset
// https://stackoverflow.com/questions/17415579/how-to-iso-8601-format-a-date-with-timezone-offset-in-javascript
// https://qiita.com/h53/items/05139982c6fd81212b08
function toISOStringWithTimezone(date: Date): string {
  const pad = function (str: string, num = 2): string {
    return ('0' + str).slice(-1 * num);
  };

  const year = date.getFullYear().toString();
  const month = pad((date.getMonth() + 1).toString());
  const day = pad(date.getDate().toString());
  const hour = pad(date.getHours().toString());
  const min = pad(date.getMinutes().toString());
  const sec = pad(date.getSeconds().toString());
  const milli = pad(date.getMilliseconds().toString(), 3);
  const tz = -date.getTimezoneOffset();

  if (tz === 0) {
    return `${year}-${month}-${day}T${hour}:${min}:${sec}.${milli}Z`;
  }

  const dif = tz >= 0 ? '+' : '-';
  const abstz = Math.abs(tz);
  const tzHour = pad(Math.floor(abstz / 60).toString());
  const tzMin = pad((abstz % 60).toString());

  return `${year}-${month}-${day}T${hour}:${min}:${sec}.${milli}${dif}${tzHour}:${tzMin}`;
}
