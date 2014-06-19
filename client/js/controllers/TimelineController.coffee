'use strict'

angular.module('sunglasses.controllers')
.controller('TimelineController', [
    '$scope',
    '$rootScope',
    '$translate',
    '$routeParams'
    'api',
    'user',
    'post',
    'photo',
    'confirm',
    ($scope, $rootScope, $translate, $routeParams, api, user, post, photo, confirm) ->
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
        $scope.apiUrl = if $routeParams.username? then 'u/' + $routeParams.username else 'timeline'
        $scope.isHome = (not $routeParams.username?)
        $scope.isSingle = ($routeParams.postid?)
        $scope.profileName = $routeParams.username
        $scope.postId = $routeParams.postid
        
        # loads a single post
        $scope.loadPost = () ->
            $scope.loading = true
            
            api(
                '/api/posts/show/' + $scope.postId,
                'GET',
                {},
                (resp) ->
                    $scope.loading = false
                    $scope.posts.push(resp.post)
                    for post in $scope.posts
                        $rootScope.relativeTime(post.created, post)
                        if post.comments?
                            for comment in post.comments
                                $rootScope.relativeTime(comment.created, comment)
                        if post.photo_url then post.photo_back = 'url(' + post.photo_url + ')'
                        if post.liked then post.className = 'liked'
            )
        
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
                '/api/' + $scope.apiUrl,
                'GET',
                params,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.loading = false
                        $scope.postCount += resp.count
                        
                        $rootScope.daTime = 5
                    
                        if loadType == 'older' and resp.posts.length < 25
                            $scope.canLoadMorePosts = false
                        else if $scope.posts.length == 0 and resp.posts.length == 25
                            $scope.canLoadMorePosts = true
                    
                        if loadType == 'older'
                            $scope.posts = $scope.posts.concat(resp.posts)
                        else
                            $scope.posts = resp.posts.concat($scope.posts)
                    
                        if not $scope.isHome
                            $rootScope.userProfile = resp.user
                    
                        for post in $scope.posts
                            $rootScope.relativeTime(post.created, post)
                            if post.comments?
                                for comment in post.comments
                                    $rootScope.relativeTime(comment.created, comment)
                            if post.photo_url then post.photo_back = 'url(' + post.photo_url + ')'
                            if post.liked then post.className = 'liked'
                    )
                , (resp) ->
                    $scope.$apply(() ->
                        $scope.loading = false
                    )
                    $rootScope.showAlert('error_code_' + resp.responseJSON.code, true, true)
            )
        
        if not $scope.isSingle
            $scope.loadPosts()
            $('.ui.dropdown').dropdown()
        
            app = document.getElementById('app')
            app.addEventListener('scroll', () ->
                
                if app.scrollHeight * 0.75 < app.scrollTop and app.scrollWidth > 760
                    $scope.$apply(() ->
                        if $scope.canLoadMorePosts then $scope.loadPosts('older')
                    )
            )
        else
            $scope.loadPost()
])
