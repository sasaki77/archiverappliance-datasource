"use strict";

function _typeof(obj) { if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.aaAddFunc = aaAddFunc;

var _lodash = _interopRequireDefault(require("lodash"));

var _jquery = _interopRequireDefault(require("jquery"));

var _core_module = _interopRequireDefault(require("app/core/core_module"));

var aafunc = _interopRequireWildcard(require("./aafunc"));

function _getRequireWildcardCache() { if (typeof WeakMap !== "function") return null; var cache = new WeakMap(); _getRequireWildcardCache = function _getRequireWildcardCache() { return cache; }; return cache; }

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } if (obj === null || _typeof(obj) !== "object" && typeof obj !== "function") { return { "default": obj }; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj["default"] = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

/** @ngInject */
function aaAddFunc($compile) {
  var inputTemplate = '<input type="text"' + ' class="gf-form-input"' + ' spellcheck="false" style="display:none"></input>';
  var buttonTemplate = '<a  class="gf-form-label tight-form-func dropdown-toggle query-part"' + ' tabindex="1" gf-dropdown="functionMenu" data-toggle="dropdown">' + '<i class="fa fa-plus"></i></a>';
  return {
    link: function link($scope, elem) {
      var categories = aafunc.getCategories();
      var allFunctions = getAllFunctionNames(categories);
      $scope.functionMenu = createFunctionDropDownMenu(categories);
      var $input = (0, _jquery["default"])(inputTemplate);
      var $button = (0, _jquery["default"])(buttonTemplate);
      $input.appendTo(elem);
      $button.appendTo(elem);
      $input.attr('data-provide', 'typeahead');
      $input.typeahead({
        source: allFunctions,
        minLength: 1,
        items: 10,
        updater: function updater(value) {
          var funcDef = aafunc.getFuncDef(value);

          if (!funcDef) {
            // try find close match
            value = value.toLowerCase();
            funcDef = _lodash["default"].find(allFunctions, function (funcName) {
              return funcName.toLowerCase().indexOf(value) === 0;
            });

            if (!funcDef) {
              return;
            }
          }

          $scope.$apply(function () {
            $scope.addFunction(funcDef);
          });
          $input.trigger('blur');
          return '';
        }
      });
      $button.click(function () {
        $button.hide();
        $input.show();
        $input.focus();
      });
      $input.keyup(function () {
        elem.toggleClass('open', $input.val() === '');
      });
      $input.blur(function () {
        // clicking the function dropdown menu won't
        // work if you remove class at once
        setTimeout(function () {
          $input.val('');
          $input.hide();
          $button.show();
          elem.removeClass('open');
        }, 200);
      });
      $compile(elem.contents())($scope);
    }
  };
}

;

_core_module["default"].directive('aaAddFunc', aaAddFunc);

function getAllFunctionNames(categories) {
  return _lodash["default"].reduce(categories, function (list, category) {
    _lodash["default"].each(category, function (func) {
      list.push(func.name);
    });

    return list;
  }, []);
}

function createFunctionDropDownMenu(categories) {
  return _lodash["default"].map(categories, function (list, category) {
    return {
      text: category,
      submenu: _lodash["default"].map(list, function (value) {
        return {
          text: value.name,
          click: "ctrl.addFunction('" + value.name + "')"
        };
      })
    };
  });
}
//# sourceMappingURL=aa-add-func.js.map
