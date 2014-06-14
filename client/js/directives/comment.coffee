'use strict'

angular.module('sunglasses')
.directive('comment', () ->
    restrict: 'E',
    templateUrl: 'templates/comment.html',
    controller: ['$scope', '$rootScope', 'api', ($scope, $rootScope, api) ->
        $scope.deleteComment = () ->
            $scope.confirm.showDialog('delete_comment_title', 'delete_comment_message', 'cancel', 'delete', () ->
                api(
                    '/api/comments/destroy/' + $scope.comment.id,
                    'DELETE',
                    confirmed: 'true',
                    (resp) ->
                        $scope.$apply(() ->
                            if resp.deleted
                                index = $scope.post.comments.indexOf($scope.comment)
                                $scope.post.comments.splice(index, 1)
                                $scope.post.comments_num -= 1
                        )
                )
            )
    ]
)