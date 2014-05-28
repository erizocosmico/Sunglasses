'use strict'

angular.module('sunglasses.controllers')
.controller('HeaderController', [
    '$scope',
    '$rootScope',
    'user',
    'api',
    ($scope, $rootScope, userService, api) ->
        $scope.userService = userService
        $scope.query = ''
        $scope.queryTimeout = null
        $scope.settingsMenuOpened = false
        $scope.notificationsMenuOpened = false
        $scope.menus = 
            settings: false
            notifications: false
        $scope.searchActive = false
        $scope.canLoadMore = false
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
                    $scope.canLoadMore = false

                    $scope.userService.search($scope.query, false, 0, 25, (resp) ->
                        $scope.$apply(() ->
                            $scope.searchResults = resp.users
                            if resp.users.length == 25
                                $scope.canLoadMore = true
                        )
                    , (resp) ->
                        console.log("Error")
                    )
                    
                    $scope.queryTimeout = null
                    $scope.searchActive = true
                )
            , 500)
        )
        
        $scope.loadMore = () ->
            $scope.canLoadMore = false
            $scope.userService.search($scope.query, false, $scope.searchResults.length, 25, (resp) ->
                $scope.$apply(() ->
                    $scope.searchResults = $scope.searchResults.concat(resp.users)
                    if resp.users.length == 25
                        $scope.canLoadMore = true
                )
            , (resp) ->
                console.log("Error")
            )
            $scope.searchActive = true
        
        $scope.sendFollowRequest = (user, isRequest) ->
            api(
                '/api/users/follow',
                'POST',
                user_to: user.id,
                (resp) ->
                    $scope.$apply(() ->
                        user.followed = true
                        user.follow_requested = !!isRequest
                    )
                , (resp) ->
                    console.log(resp)
            )
            
        $scope.unfollow = (user) ->
            api(
                '/api/users/unfollow',
                'DELETE',
                user_to: user.id,
                (resp) ->
                    $scope.$apply(() ->
                        user.followed = false
                    )
                , (resp) ->
                    console.log(resp)
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
