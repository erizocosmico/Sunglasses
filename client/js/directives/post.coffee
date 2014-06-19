'use strict'

angular.module('sunglasses')
.directive('post', () ->
    restrict: 'E',
    replace: true,
    templateUrl: 'templates/post.html',
    controller: ['$scope', '$rootScope', 'api', ($scope, $rootScope, api) ->
        $scope.post.commentsDirty = 0
        $scope.privacyOpened = false
        $scope.privacySelected = $scope.post.privacy.privacy_type
        
        if not $scope.post.comments? then $scope.post.comments = []
        
        # toggles the privacy selector
        $scope.togglePrivacy = () ->
            $scope.privacyOpened = not $scope.privacyOpened

        # delete a post
        $scope.deletePost = () ->
            $scope.confirm.showDialog('delete_post_title', 'delete_post_message', 'cancel', 'delete', () ->
                api(
                    '/api/posts/destroy/' + $scope.post.id,
                    'DELETE',
                    null,
                    (resp) ->
                        $scope.$apply(() ->
                            index = $scope.posts.indexOf($scope.post)
                            $scope.posts.splice(index, 1)
                        )
                )
            )
                
        # likes a post
        $scope.likePost = () ->
            api(
                '/api/posts/like/' + $scope.post.id,
                'PUT',
                null,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.post.liked = resp.liked
                        $scope.post.likes += if resp.liked then 1 else -1
                    
                        if resp.liked
                            $scope.post.className = 'liked animated bounceIn'
                        else
                            $scope.post.className = ''
                    )
            )
            
        # changes the privacy of the post
        $scope.changePostPrivacy = (type) ->
            api(
                '/api/posts/change_privacy/' + $scope.post.id,
                'PUT',
                privacy_type: type,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.post.privacy.privacy_type = type
                        $scope.privacySelected = type
                    )
                    $rootScope.showAlert('privacy_updated', true)
            )
            
        $scope.commentPost = () ->
            commentArea = $('#comment-area-' + $scope.post.id).focus()
            return
            
        $scope.loadMoreComments = () ->
            api(
                '/api/comments/for_post/' + $scope.post.id,
                'GET',
                older_than: $scope.post.comments[$scope.post.comments.length - 1 - $scope.post.commentsDirty].created,
                (resp) ->
                    $scope.$apply(() ->
                        for c in resp.comments
                            found = false
                            for cTmp in $scope.post.comments
                                if cTmp.id == c.id
                                    found = true
                                    break
                                    
                            if not found
                                $rootScope.relativeTime(c.created, c)
                                $scope.post.comments.push(c)
                        $scope.post.commentsDirty = 0
                    )
            )
    ]
)