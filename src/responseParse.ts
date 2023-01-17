import _ from 'lodash';
import {
    MutableDataFrame,
    FieldType,
} from '@grafana/data';

import { getToScalarFuncs } from './aafunc';
import {
    TargetQuery,
    AADataQueryData,
    AADataQueryResponse,
    isNumberArray,
} from './types';

export function responseParse(responses: AADataQueryResponse[], target: TargetQuery) {
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

    // Except for raw operator or extrapolation is disabled
    if ((target.operator !== 'raw' && target.interval !== '') || target.options.disableExtrapol === 'true') {
        return Promise.resolve(dataFrames);
    }

    // Extrapolation for raw operator
    const to_msec = target.to.getTime();
    const extrapolationDataFrames = _.map(dataFrames, (dataframe) => {
        const latestval = dataframe.get(dataframe.length - 1);
        const addval = { ...latestval, time: to_msec };

        dataframe.add(addval);

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
        return new MutableDataFrame();
    }

    const frames = _.map(toScalarFuncs, (func) => {
        const values = _.map(targetRes.data, (datapoint) => func.func(datapoint.val));
        const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
        const frame = new MutableDataFrame({
            refId: target.refId,
            name: targetRes.meta.name,
            fields: [
                { name: 'time', type: FieldType.time, values: times },
                {
                    name: 'value',
                    type: FieldType.number,
                    values: values,
                    config: { displayName: `${targetRes.meta.name} (${func.label})` },
                },
            ],
        });
        return frame;
    });

    return frames;
}

function parseArrayResponse(targetRes: AADataQueryData, target: TargetQuery) {
    let fields;
    if (target.options.arrayFormat == "dt-space") {
        fields = makeDtSpaceArrayFields(targetRes);
    } else if (target.options.arrayFormat == "index") {
        fields = makeIndexArrayFields(targetRes);
    } else {
        fields = makeTimeseriesArrayFields(targetRes);
    }

    if (fields.length == 0) {
        return new MutableDataFrame();
    }

    const frame = new MutableDataFrame({
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

    const targetData = targetRes.data

    const field_val = _.reduce(
        targetData,
        (fields, data, i) => {
            fields["vals"] = fields["vals"].concat(data.val);

            const len = data.val.length;
            for (let i = 0; i < len; i++) {
                const date = data.millis + i;
                fields["times"].push(date)
            }

            return fields;
        },
        { times: [], vals: [] } as { times: number[]; vals: number[] }
    );

    const fields = [
        { name: 'time', type: FieldType.time, values: field_val["times"] },
        { name: targetRes.meta.name, type: FieldType.number, values: field_val["vals"] }
    ];

    return fields;
}

function makeIndexArrayFields(targetRes: AADataQueryData) {
    // Type check for columnValues
    if (!isNumberArray(targetRes)) {
        return [];
    }

    const targetData = targetRes.data

    const len = targetData[0].val.length;
    let numbers = []
    for (let i = 0; i < len; i++) {
        numbers.push(i)
    }

    const fields = [{ name: 'index', type: FieldType.number, values: numbers }];

    _.reduce(
        targetData,
        (fields, data, i) => {
            const date = new Date(data.millis);
            const val = data.val.length >= len ? data.val.slice(0, len) : data.val;
            const field = {
                name: date.toISOString(),
                type: FieldType.number,
                values: val,
            };
            fields.push(field);
            return fields;
        },
        fields
    );

    return fields;
}

function makeTimeseriesArrayFields(targetRes: AADataQueryData) {
    // Type check for columnValues
    if (!isNumberArray(targetRes)) {
        return [];
    }

    const columnValues = _.map(targetRes.data, (datapoint) => datapoint.val);

    const rowValues = _.unzip(columnValues);
    const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
    const fields = [{ name: 'time', type: FieldType.time, values: times }];

    // Add fields for each waveform elements
    _.reduce(
        rowValues,
        (fields, val, i) => {
            const field = {
                name: `${targetRes.meta.name}[${i}]`,
                type: FieldType.number,
                values: val,
            };
            fields.push(field);
            return fields;
        },
        fields
    );

    return fields;
}

function parseScalarResponse(targetRes: AADataQueryData, target: TargetQuery): MutableDataFrame {
    const values = _.map(targetRes.data, (datapoint) => datapoint.val);
    const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
    const frame = new MutableDataFrame({
        refId: target.refId,
        name: targetRes.meta.name,
        fields: [
            { name: 'time', type: FieldType.time, values: times },
            { name: 'value', type: FieldType.number, values: values, config: { displayName: targetRes.meta.name } },
        ],
    });
    return frame;
}
