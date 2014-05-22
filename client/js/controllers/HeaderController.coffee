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
        $scope.menus = 
            settings: false
            notifications: false
        
        # Perform a search
        $scope.$watch('query', () ->
            if $scope.queryTimeout?
                window.clearTimeout($scope.queryTimeout)
            
            $scope.queryTimeout = window.setTimeout(() ->
                console.log 'searching: ' + $scope.query
                $scope.queryTimeout = null
            , 500)
        )

        $scope.toggleMenu = (menuType, closeCallback) ->
            otherMenu = if menuType == 'settings' then 'notifications' else 'settings'
            if $scope.menus[otherMenu]
                document.getElementById('#' + otherMenu + '-menu').className += ' hidden'
            
            menu = document.getElementById(menuType + '-menu')
            if $scope.menus[menuType]
                $scope.menus[menuType] = false
                $rootScope.animateElem(menu, 'bounceOutUp', () ->
                    menu.className = 'hidden ng-scope'
                )
            else
                $rootScope.animateElem(menu, 'bounceInDown')
                $scope.menus[menuType] = true
               
            # CoffeeScript bad habit 
            return
])
