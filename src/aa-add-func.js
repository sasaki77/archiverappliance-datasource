import _ from 'lodash';
import $ from 'jquery';

import coreModule from 'app/core/core_module';
import * as aafunc from './aafunc';

function getAllFunctionNames(categories) {
  return _.reduce(categories, (list, category) => {
    _.each(category, (func) => list.push(func.name));
    return list;
  }, []);
}

function createFunctionDropDownMenu(categories) {
  return _.map(categories, (list, category) => (
    {
      text: category,
      submenu: _.map(list, (value) => (
        {
          text: value.name,
          click: `ctrl.addFunction('${value.name}')`,
        }
      )),
    }
  ));
}

/** @ngInject */
export function aaAddFunc($compile) {
  const inputTemplate = '<input type="text"'
                        + ' class="gf-form-input"'
                        + ' spellcheck="false" style="display:none"></input>';

  const buttonTemplate = '<a  class="gf-form-label tight-form-func dropdown-toggle query-part"'
                         + ' tabindex="1" gf-dropdown="functionMenu" data-toggle="dropdown">'
                         + '<i class="fa fa-plus"></i></a>';


  return {
    link: ($scope, elem) => {
      const categories = aafunc.getCategories();
      const allFunctions = getAllFunctionNames(categories);

      $scope.functionMenu = createFunctionDropDownMenu(categories);

      const $input = $(inputTemplate);
      const $button = $(buttonTemplate);
      $input.appendTo(elem);
      $button.appendTo(elem);

      $input.attr('data-provide', 'typeahead');
      $input.typeahead({
        source: allFunctions,
        minLength: 1,
        items: 10,
        updater: (value) => {
          let funcDef = aafunc.getFuncDef(value);
          if (!funcDef) {
            // try find close match
            const lowerValue = value.toLowerCase();
            funcDef = _.find(allFunctions, (funcName) => (
              funcName.toLowerCase().indexOf(lowerValue) === 0
            ));

            if (!funcDef) { return undefined; }
          }

          $scope.$apply(() => {
            $scope.addFunction(funcDef);
          });

          $input.trigger('blur');
          return '';
        },
      });

      $button.click(() => {
        $button.hide();
        $input.show();
        $input.focus();
      });

      $input.keyup(() => {
        elem.toggleClass('open', $input.val() === '');
      });

      $input.blur(() => {
        // clicking the function dropdown menu won't
        // work if you remove class at once
        setTimeout(() => {
          $input.val('');
          $input.hide();
          $button.show();
          elem.removeClass('open');
        }, 200);
      });

      $compile(elem.contents())($scope);
    },
  };
}

coreModule.directive('aaAddFunc', aaAddFunc);
