"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.FunctionEditor = void 0;

var _react = _interopRequireDefault(require("react"));

var _ui = require("@grafana/ui");

var _FunctionEditorControls = require("./FunctionEditorControls");

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

function _typeof(obj) { if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

function _extends() { _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; }; return _extends.apply(this, arguments); }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

function _possibleConstructorReturn(self, call) { if (call && (_typeof(call) === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function"); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, writable: true, configurable: true } }); if (superClass) _setPrototypeOf(subClass, superClass); }

function _setPrototypeOf(o, p) { _setPrototypeOf = Object.setPrototypeOf || function _setPrototypeOf(o, p) { o.__proto__ = p; return o; }; return _setPrototypeOf(o, p); }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var FunctionEditor =
/*#__PURE__*/
function (_React$PureComponent) {
  _inherits(FunctionEditor, _React$PureComponent);

  function FunctionEditor(props) {
    var _this;

    _classCallCheck(this, FunctionEditor);

    _this = _possibleConstructorReturn(this, _getPrototypeOf(FunctionEditor).call(this, props));

    _defineProperty(_assertThisInitialized(_this), "triggerRef", _react["default"].createRef());

    _defineProperty(_assertThisInitialized(_this), "renderContent", function (_ref) {
      var updatePopperPosition = _ref.updatePopperPosition;
      var _this$props = _this.props,
          _onMoveLeft = _this$props.onMoveLeft,
          _onMoveRight = _this$props.onMoveRight,
          _this$props$func$def = _this$props.func.def,
          name = _this$props$func$def.name,
          description = _this$props$func$def.description;
      var showingDescription = _this.state.showingDescription;

      if (showingDescription) {
        return _react["default"].createElement("div", {
          style: {
            overflow: 'auto',
            maxHeight: '30rem',
            textAlign: 'left',
            fontWeight: 'normal'
          }
        }, _react["default"].createElement("h4", {
          style: {
            color: 'white'
          }
        }, " ", name, " "), _react["default"].createElement("div", null, description));
      }

      return _react["default"].createElement(_FunctionEditorControls.FunctionEditorControls, _extends({}, _this.props, {
        onMoveLeft: function onMoveLeft() {
          _onMoveLeft(_this.props.func);

          updatePopperPosition();
        },
        onMoveRight: function onMoveRight() {
          _onMoveRight(_this.props.func);

          updatePopperPosition();
        },
        onDescriptionShow: function onDescriptionShow() {
          _this.setState({
            showingDescription: true
          }, function () {
            updatePopperPosition();
          });
        }
      }));
    });

    _this.state = {
      showingDescription: false
    };
    return _this;
  }

  _createClass(FunctionEditor, [{
    key: "render",
    value: function render() {
      var _this2 = this;

      return _react["default"].createElement(_ui.PopoverController, {
        content: this.renderContent,
        placement: "top",
        hideAfter: 300
      }, function (showPopper, hidePopper, popperProps) {
        return _react["default"].createElement(_react["default"].Fragment, null, _this2.triggerRef && _react["default"].createElement(_ui.Popover, _extends({}, popperProps, {
          referenceElement: _this2.triggerRef.current,
          wrapperClassName: "popper",
          className: "popper__background",
          onMouseLeave: function onMouseLeave() {
            _this2.setState({
              showingDescription: false
            });

            hidePopper();
          },
          onMouseEnter: showPopper,
          renderArrow: function renderArrow(_ref2) {
            var arrowProps = _ref2.arrowProps,
                placement = _ref2.placement;
            return _react["default"].createElement("div", _extends({
              className: "popper__arrow",
              "data-placement": placement
            }, arrowProps));
          }
        })), _react["default"].createElement("span", {
          ref: _this2.triggerRef,
          onClick: popperProps.show ? hidePopper : showPopper,
          onMouseLeave: function onMouseLeave() {
            hidePopper();

            _this2.setState({
              showingDescription: false
            });
          },
          style: {
            cursor: 'pointer'
          }
        }, _this2.props.func.def.name));
      });
    }
  }]);

  return FunctionEditor;
}(_react["default"].PureComponent);

exports.FunctionEditor = FunctionEditor;
//# sourceMappingURL=FunctionEditor.js.map
