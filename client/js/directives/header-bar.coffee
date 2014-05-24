'use strict'

angular.module('sunglasses')
.directive('headerBar', () ->
    restrict: 'E',
    # Avoid cache, TODO: remove for production
    templateUrl: 'templates/header.html?' + new Date().getTime(),
    controller: 'HeaderController'
)