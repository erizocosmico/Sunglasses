'use strict'

angular.module('sunglasses.controllers')
.controller('LandingController', ['$rootScope', ($rootScope) -> $rootScope.title = 'sunglasses'])
.controller('HomeController', [
    '$scope',
    '$rootScope',
    '$translate',
    'api',
    'user',
    'post',
    'photo',
    'confirm',
    ($scope, $rootScope, $translate, api, user, post, photo, confirm) ->
        # number of posts retrieved
        $scope.postCount = 0
        # array of the previously retrieved posts
        $scope.posts = []
        # loading shows the loading dialog when new posts are being loaded
        $scope.loading = true
        $scope.userService = user
        $scope.postService = post
        $scope.photoService = photo
        $scope.confirm = confirm
        $scope.canLoadMorePosts = false
        
        # load more posts, uses $scope.postCount to automatically manage pagination
        $scope.loadPosts = (loadType) ->
            params = {}   
            $scope.loading = true
                
            if loadType?
                switch loadType
                    when 'older'
                        params.older_than = $scope.posts[$scope.posts.length - 1].created
                    else
                        params.newer_than = $scope.posts[0].created

            api(
                '/api/timeline',
                'GET',
                params,
                (resp) ->
                    $scope.loading = false
                    $scope.postCount += resp.count
                    if loadType == 'older'
                        $scope.posts = $scope.posts.concat(resp.posts)
                    else
                        $scope.posts = resp.posts.concat($scope.posts)
                    
                    $scope.canLoadMorePosts = (not loadType? or loadType == 'older') and resp.count == 25
                    
                    for post in $scope.posts
                        $rootScope.relativeTime(post.created, post)
                        if post.comments?
                            for comment in post.comments
                                $rootScope.relativeTime(comment.created, comment)
                        if post.photo_url then post.photo_back = 'url(' + post.photo_url + ')'
                        if post.liked then post.className = 'liked'
                , (resp) ->
                    $scope.loading = false
                    $rootScope.showMsg('error_code_' + resp.responseJSON.code, 'timeline-error')
            )
                
        $scope.loadPosts()
        $('.ui.dropdown').dropdown()
])
