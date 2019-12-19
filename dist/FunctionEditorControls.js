"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.FunctionEditorControls = void 0;

var _react = _interopRequireDefault(require("react"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var FunctionHelpButton = function FunctionHelpButton(props) {
  if (props.description) {
    return _react.default.createElement("span", {
      className: "pointer fa fa-question-circle",
      onClick: props.onDescriptionShow
    });
  }

  return _react.default.createElement("span", {
    className: "pointer fa fa-question-circle",
    onClick: function onClick() {
      window.open("https://sasaki77.github.io/archiverappliance-datasource/functions.html#".concat(props.name), '_blank');
    }
  });
};

var FunctionEditorControls = function FunctionEditorControls(props) {
  var func = props.func,
      onMoveLeft = props.onMoveLeft,
      onMoveRight = props.onMoveRight,
      onRemove = props.onRemove,
      onDescriptionShow = props.onDescriptionShow;
  return _react.default.createElement("div", {
    style: {
      display: 'flex',
      width: '60px',
      justifyContent: 'space-between'
    }
  }, _react.default.createElement("span", {
    className: "pointer fa fa-arrow-left",
    onClick: function onClick() {
      return onMoveLeft(func);
    }
  }), _react.default.createElement(FunctionHelpButton, {
    name: func.def.name,
    description: func.def.description,
    onDescriptionShow: onDescriptionShow
  }), _react.default.createElement("span", {
    className: "pointer fa fa-remove",
    onClick: function onClick() {
      return onRemove(func);
    }
  }), _react.default.createElement("span", {
    className: "pointer fa fa-arrow-right",
    onClick: function onClick() {
      return onMoveRight(func);
    }
  }));
};

exports.FunctionEditorControls = FunctionEditorControls;
//# sourceMappingURL=FunctionEditorControls.js.map
