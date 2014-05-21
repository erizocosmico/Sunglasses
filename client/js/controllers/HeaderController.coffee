'use strict'

angular.module('sunglasses.controllers')
.controller('HeaderController', [
    '$scope',
    '$rootScope',
    'api',
    ($scope, $rootScope, api) ->
        $scope.query = ''
        $scope.queryTimeout = null
        $scope.settingsMenuOpened = false
        $scope.notificationsMenuOpened = false
        
        # Perform a search 
        $scope.$watch('query', () ->
            if $scope.queryTimeout?
                window.clearTimeout($scope.queryTimeout)
            
            $scope.queryTimeout = window.setTimeout(() ->
                console.log 'searching: ' + $scope.query
                $scope.queryTimeout = null
            , 500)
        )
        
        # TODO: Rewrite
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
