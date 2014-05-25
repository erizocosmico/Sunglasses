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
            
        # shows an error or a success message
        $rootScope.showMsg = (text, field, success) ->
            $translate(text).then (msg) ->
                document.getElementById(field).innerHTML = msg
                $rootScope.displayError(field, success)
        
        # load more posts, uses $scope.postCount to automatically manage pagination
        $scope.loadPosts = (withoutOffset) ->
            params = 
                count: 25,
                offset: if withoutOffset? then 0 else $scope.postCount
                
            $scope.loading = true
                
            if $scope.posts.length > 0 and withoutOffset?
                params.newer_than = $scope.posts[0].created
                params.count = 50

            api(
                '/api/timeline',
                'GET',
                params,
                (resp) ->
                    $scope.loading = false
                    $scope.postCount += resp.count
                    $scope.posts = resp.posts.concat($scope.posts)
                    
                    for post in $scope.posts
                        $rootScope.relativeTime(post.created, post)
                        if post.liked then post.className = 'liked'
                , (resp) ->
                    $scope.loading = false
                    $rootScope.showMsg('error_code_' + resp.responseJSON.code, 'timeline-error')
            )
            
        # delete a post
        $scope.deletePost = (post) ->
            $scope.confirm.showDialog('delete_post_title', 'delete_post_message', 'cancel', 'delete', () ->
                api(
                    '/api/posts/destroy/' + post.id,
                    'DELETE',
                    null,
                    (resp) ->
                        $scope.$apply(() ->
                            index = $scope.posts.indexOf(post)
                            $scope.posts.splice(index, 1)
                        )
                    , (resp) ->
                        # TODO: General error handling
                        console.log(resp)
                )
            )
                
        # likes a post
        $scope.likePost = (index) ->
            api(
                '/api/posts/like/' + $scope.posts[index].id,
                'PUT',
                null,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.posts[index].liked = resp.liked
                        $scope.posts[index].likes += if resp.liked then 1 else -1
                    
                        if resp.liked
                            $scope.posts[index].className = 'liked animated bounceIn'
                        else
                            $scope.posts[index].className = ''
                    )
                , (resp) ->
                    # TODO default error popup
                    $rootScope.showMsg('error_code_' + resp.responseJSON.code, 'post-error')
            )
                
        $scope.loadPosts()
        $('.ui.dropdown').dropdown()
])
