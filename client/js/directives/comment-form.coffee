'use strict'

angular.module('sunglasses')
.directive('commentForm', () ->
    restrict: 'E',
    templateUrl: 'templates/comment-form.html',
    controller: ['$scope', '$rootScope', 'api', ($scope, $rootScope, api) ->
        $scope.commentText = ''
        
        $scope.postComment = () ->
            if $scope.commentText.trim().length <= 0 then return
              
            api(
                '/api/comments/create',
                'POST',
                post_id: $scope.post.id,
                comment_text: $scope.commentText.trim(),
                (resp) ->
                    if not resp.error
                        $scope.$apply(() ->
                            $scope.post.comments.push(resp.comment)
                            $scope.post.comments_num += 1
                        )
                , (resp) ->
                    console.log(resp)
            )
    ]
)