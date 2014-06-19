'use strict'

angular.module('sunglasses')
.directive('timeline', () ->
    restrict: 'E',
    replace: true,
    templateUrl: 'templates/timeline.html',
    controller: 'TimelineController',
)