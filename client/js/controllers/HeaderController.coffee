'use strict'

angular.module('sunglasses.controllers')
.controller('HeaderController', [
    '$scope',
    '$rootScope',
    'api',
    ($scope, $rootScope, api) ->
        # Avoid template caching
        # TODO: Remove for production
        $scope.templateUrl = 'templates/header.html?t=' + new Date().getTime()
        $scope.query = ''
])
