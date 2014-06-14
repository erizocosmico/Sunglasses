'use strict'

angular.module('sunglasses')
.directive('headerBar', () ->
    restrict: 'E',
    templateUrl: 'templates/header.html',
    controller: 'HeaderController'
)