'use strict'

angular.module('sunglasses')
.directive('headerBar', () ->
    restrict: 'E',
    replace: true,
    templateUrl: 'templates/header.html',
    controller: 'HeaderController'
)