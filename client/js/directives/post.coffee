'use strict'

angular.module('sunglasses')
.directive('post', () ->
    restrict: 'E',
    templateUrl: 'templates/post.html',
    controller: ['$scope', '$rootScope', 'api', ($scope, $rootScope, api) ->
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
                    , (resp) ->
                        # TODO: General error handling
                        console.log(resp)
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
                , (resp) ->
                    $rootScope.showMsg('error_code_' + resp.responseJSON.code, 'post-error')
            )
    ]
)