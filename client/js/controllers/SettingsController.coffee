'use strict'

angular.module('sunglasses.controllers')
.controller('SettingsController', [
    '$scope',
    '$rootScope',
    'api',
    ($scope, $rootScope, api) ->
        $rootScope.title = 'settings'
        
        $scope.settings = userData.settings
        console.log $scope.settings.invisible
        
        $scope.activeSection = 'account_details'
        $scope.setActiveSection = (section) ->
            $scope.activeSection = section
            
        (() ->
            $('.ui.checkbox').checkbox()
        )()
])
