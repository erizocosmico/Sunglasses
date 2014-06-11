'use strict'

angular.module('sunglasses.controllers')
.controller('ProfileController', [
    '$routeParams',
    '$rootScope',
    '$scope',
    'user',
    'api',
    ($routeParams, $rootScope, $scope, user, api) ->
        $scope.userService = user
        $rootScope.userProfile = {}
        $scope.infoVisible = false
        
        $scope.toggleInfo = () ->
            $scope.infoVisible = !$scope.infoVisible
        
        # Get user info
])