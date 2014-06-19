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
        $scope.showTimeline = true
        $scope.follows = []
        $scope.followType = ''
        $scope.loadMoreFollows = false
        
        $scope.toggleInfo = () ->
            $scope.infoVisible = !$scope.infoVisible
            
        $scope.loadFollows = (type) ->
            if $scope.userProfile.id != $rootScope.userData.id then return
            
            $scope.showTimeline = false
            followType = if type == 'followers' then 'followers' else 'following'
            if $scope.followType != followType then $scope.follows = []
            $scope.followType = followType
            $scope.loadMoreFollows = false

            data =
                count: 25,
                offset: $scope.follows.length
            
            api(
                '/api/users/' + followType,
                'GET',
                data,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.loadMoreFollows = (resp.count == 25)
                        $scope.follows = $scope.follows.concat(resp[followType])
                        
                        for f in resp[followType]
                            $rootScope.relativeTime(f.time, f.user)
                    )
            )
])