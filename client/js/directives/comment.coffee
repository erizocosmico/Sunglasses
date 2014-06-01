'use strict'

angular.module('sunglasses')
.directive('comment', () ->
    restrict: 'E',
    templateUrl: 'templates/comment.html',
    controller: ['$scope', '$rootScope', 'api', ($scope, $rootScope, api) ->
        #stuff
    ]
)