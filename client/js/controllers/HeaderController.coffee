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
        
        $scope.toggleSettingsMenu = () ->
            menu = document.getElementById('settings-menu')
            if menu.className.indexOf('hidden ng-scope') != -1
                $rootScope.animateElem(menu, 'bounceInDown')
            else
                $rootScope.animateElem(menu, 'bounceOutUp', () ->
                    menu.className = 'hidden ng-scope'
                )
            return
])
