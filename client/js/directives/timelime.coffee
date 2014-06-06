'use strict'

angular.module('sunglasses')
.directive('timeline', () ->
    restrict: 'E',
    templateUrl: 'templates/timeline.html',
    controller: 'TimelineController'
)