"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.FunctionEditorControls = void 0;

var _react = _interopRequireDefault(require("react"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

var FunctionHelpButton = function FunctionHelpButton(props) {
  if (props.description) {
    return <span className="pointer fa fa-question-circle" onClick={props.onDescriptionShow} />;
  }

  return <span className="pointer fa fa-question-circle" onClick={function () {
    window.open('https://github.com/sasaki77/archiverappliance-datasource', '_blank');
  }} />;
};

var FunctionEditorControls = function FunctionEditorControls(props) {
  var func = props.func,
      onMoveLeft = props.onMoveLeft,
      onMoveRight = props.onMoveRight,
      onRemove = props.onRemove,
      onDescriptionShow = props.onDescriptionShow;
  return <div style={{
    display: 'flex',
    width: '60px',
    justifyContent: 'space-between'
  }}>
      <span className="pointer fa fa-arrow-left" onClick={function () {
      return onMoveLeft(func);
    }} />
      <FunctionHelpButton name={func.def.name} description={func.def.description} onDescriptionShow={onDescriptionShow} />
      <span className="pointer fa fa-remove" onClick={function () {
      return onRemove(func);
    }} />
      <span className="pointer fa fa-arrow-right" onClick={function () {
      return onMoveRight(func);
    }} />
    </div>;
};

exports.FunctionEditorControls = FunctionEditorControls;
//# sourceMappingURL=FunctionEditorControls.js.map
