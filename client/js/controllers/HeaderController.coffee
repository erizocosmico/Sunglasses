'use strict'

angular.module('sunglasses.controllers')
.controller('HeaderController', [
    '$scope',
    '$rootScope',
    'user',
    ($scope, $rootScope, userService) ->
        $scope.userService = userService
        $scope.query = ''
        $scope.queryTimeout = null
        $scope.settingsMenuOpened = false
        $scope.notificationsMenuOpened = false
        $scope.menus = 
            settings: false
            notifications: false
        $scope.searchActive = false
        $scope.searchResults = []
        
        # Perform a search
        $scope.$watch('query', () ->
            if $scope.queryTimeout?
                window.clearTimeout($scope.queryTimeout)
            
            $scope.queryTimeout = window.setTimeout(() ->
                $scope.$apply(() ->
                    if $scope.query.trim() == ''
                        $scope.searchActive = false
                        return

                    $scope.userService.search($scope.query, false, 0, 25, (resp) ->
                        $scope.$apply(() ->
                            $scope.searchResults = resp.users
                        )
                    , (resp) ->
                        console.log("Error")
                    )
                    
                    $scope.queryTimeout = null
                    $scope.searchActive = true
                )
            , 500)
        )

        # Toggles a menu
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
