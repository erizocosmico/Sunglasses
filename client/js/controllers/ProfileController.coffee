'use strict'

angular.module('sunglasses.controllers')
.controller('ProfileController', [
    '$routeParams',
    '$rootScope',
    '$scope',
    'user',
    'api',
    ($routeParams, $rootScope, $scope, userService, api) ->
        $rootScope.title = 'sunglasses'
        $scope.userService = userService
        $rootScope.userProfile = {}
        $scope.infoVisible = false
        
        $scope.toggleInfo = () ->
            $scope.infoVisible = !$scope.infoVisible
])