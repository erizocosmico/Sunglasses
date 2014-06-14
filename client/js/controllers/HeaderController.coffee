'use strict'

angular.module('sunglasses.controllers')
.controller('HeaderController', [
    '$scope',
    '$rootScope',
    '$timeout',
    'user',
    'api',
    ($scope, $rootScope, $timeout, userService, api) ->
        # Services
        $scope.userService = userService
        
        # Search vars
        $scope.query = ''
        $scope.queryTimeout = null
        $scope.searchActive = false
        $scope.canLoadMore = false
        $scope.searchResults = []
        
        # Menu vars
        $scope.settingsMenuOpened = false
        $scope.notificationsMenuOpened = false
        $scope.menus = 
            settings: false
            notifications: false
            
        # Notification vars
        $scope.notifications = []
        $scope.canLoadMoreNotifications = false
        $scope.unreadCount = 0
        
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
    
        # Load more search results
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

        # Toggles a menu
        $scope.toggleMenu = (menuType, closeCallback) ->
            otherMenu = if menuType == 'settings' then 'notifications' else 'settings'
            if $scope.menus[otherMenu]
                $scope.menus[otherMenu] = false
                document.getElementById(otherMenu + '-menu').className += ' hidden'
            
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
            
        # Load notifications
        $scope.loadNotifications = () ->
            api(
                '/api/notifications/list',
                'GET',
                count: 25,
                offset: $scope.notifications.length,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.notifications = $scope.notifications.concat(resp.notifications)
                        $scope.canLoadMoreNotifications = resp.notifications.length == 25
                        
                        for n in resp.notifications
                            $rootScope.relativeTime(n.time, n)
                            if not n.read
                                $scope.unreadCount += 1
                    )
                , (resp) ->
                    console.log(resp)
            )
            
        # Load more notifications every 3 minutes
        notificationInterval = () ->
            $timeout(() ->
                $scope.loadNotifications()
                notificationInterval()
            , 180000)
            
        notificationInterval()
        
        document.getElementsByClassName('header-overlay')[0].addEventListener('click', (e) ->
            if e.target.id == 'header-overlay'
                $scope.$apply(() ->
                    $scope.query = ''
                )
        )
])
