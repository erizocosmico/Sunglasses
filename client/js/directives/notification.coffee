'use strict'

angular.module('sunglasses')
.directive('notification', ['api', (api) ->
    restrict: 'E',
    template: '<div ng-switch="notification.notification_type">
                    <div ng-switch-when="1">
                        <a ng-href="#/u/{{ notification.user_action.username }}">{{ userService.getUsername(notification.user_action) }}</a> {{ \'has_sent_follow_request\' | translate }}
                        <div class="block centered">
                            <p ng-show="replied && accepted">{{ \'follow_accepted\' | translate }}</p>
                            <p ng-show="replied && !accepted">{{ \'follow_declined\' | translate }}</p>
                            <button class="btn small-btn" ng-show="!replied" ng-click="replyFollowRequest(\'yes\')">{{ \'accept\' | translate }}</button>
                            <button class="btn small-btn btn-white" ng-show="!replied" ng-click="replyFollowRequest(\'no\')">{{ \'decline\' | translate }}</button>
                        </div>
                    </div>
                    <div ng-switch-when="2">
                        Follow request accepted
                    </div>
                    <div ng-switch-when="3">
                        Followed
                    </div>
                    <div ng-switch-when="4">
                        Like
                    </div>
                    <div ng-switch-when="5">
                        Comment
                    </div>
                    <div ng-switch-when="6">
                        Wall post
                    </div>
                </div>
    ',
    link: (scope, elem, attrs) ->
        scope.replied = false
        scope.accepted = false
    , controller: ['$scope', ($scope) ->
        $scope.replyFollowRequest = (accept) ->
            api(
                '/api/users/reply_follow_request',
                'POST',
                accept: if accept == 'yes' then accept else 'no',
                from_user: $scope.notification.user_action.id,
                to_user: $scope.notification.user_id,
                (resp) ->
                    $scope.$apply(() ->
                        $scope.replied = true
                        $scope.accepted = accept == 'yes'
                        $scope.notification.read = true
                    )
                , (resp) ->
                    console.log(resp)
            )
            
        $scope.performAction = () ->
            console.log "Perform action"
    ]
])