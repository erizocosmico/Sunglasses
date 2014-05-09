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
        $scope.settingsMenuOpened = false
        $scope.notificationsMenuOpened = false
        
        $scope.toggleMenu = (menuType, closeCallback) ->
            m = menuType + '-menu'
            menu = document.getElementById(m)
            if menu.className.indexOf('hidden') != -1
                if menuType == 'settings'
                    $scope.settingsMenuOpened = true
                    
                    if $scope.notificationsMenuOpened
                        $scope.toggleMenu('notifications', () ->
                            $rootScope.animateElem(menu, 'bounceInDown')
                        )
                        return
                else
                    $scope.notificationsMenuOpened = true
                    
                    if $scope.settingsMenuOpened
                        $scope.toggleMenu('settings', () ->
                            $rootScope.animateElem(menu, 'bounceInDown')
                        )
                        return
                $rootScope.animateElem(menu, 'bounceInDown')
            else
                if menuType == 'settings'
                    $scope.settingsMenuOpened = false
                else
                    $scope.notificationsMenuOpened = false
                $rootScope.animateElem(menu, 'bounceOutUp', () ->
                    menu.className = 'hidden ng-scope'
                        
                    if closeCallback? then closeCallback()
                )
            return
])
