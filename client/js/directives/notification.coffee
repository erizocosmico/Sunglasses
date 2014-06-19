'use strict'

angular.module('sunglasses')
.directive('notification', () ->
    restrict: 'E',
    template: '<div ng-switch="notification.notification_type" ng-click="performAction()" ng-class="{\'cursor-pointer\': notification.notification_type > 1}">
                    <div class="user-data">
                        <div class="avatar" ng-hide="userService.getAvatarThumb(notification.user_action) == \'\'">
                            <img ng-src="{{ userService.getAvatarThumb(notification.user_action) }}" alt="{{ userService.getUsername(notification.user_action) }}">
                        </div>
                        <div class="default-avatar" ng-hide="userService.getAvatarThumb(notification.user_action) != \'\'">
                            <span class="ion ion-android-contact"></span>
                        </div>
                    </div>
                    <div ng-switch-when="1">
                        <a ng-href="#/u/{{ notification.user_action.username }}">{{ userService.getUsername(notification.user_action) }}</a> {{ \'has_sent_follow_request\' | translate }}
                        <div class="block centered clear">
                            <p ng-show="replied && accepted">{{ \'follow_accepted\' | translate }}</p>
                            <p ng-show="replied && !accepted">{{ \'follow_declined\' | translate }}</p>
                            <button class="btn small-btn" ng-show="!replied" ng-click="replyFollowRequest(\'yes\')">{{ \'accept\' | translate }}</button>
                            <button class="btn small-btn btn-white" ng-show="!replied" ng-click="replyFollowRequest(\'no\')">{{ \'decline\' | translate }}</button>
                        </div>
                    </div>
                    <div ng-switch-when="2">
                        <a>{{ userService.getUsername(notification.user_action) }}</a> {{ \'has_accepted_your_follow_request\' | translate }}
                    </div>
                    <div ng-switch-when="3">
                        <a>{{ userService.getUsername(notification.user_action) }}</a> {{ \'has_followed_you\' | translate }}
                    </div>
                    <div ng-switch-when="4">
                        <a>{{ userService.getUsername(notification.user_action) }}</a> {{ \'has_liked_your_post\' | translate }}
                    </div>
                    <div ng-switch-when="5">
                        <a>{{ userService.getUsername(notification.user_action) }}</a> {{ \'has_commented_your_post\' | translate }}
                    </div>
                    <div ng-switch-when="6">
                        Wall post
                    </div>
                    <span class="time" translate="time_format" translate-value-unit="{{ notification.timeUnit | translate }}" translate-value-num="{{ notification.timeNumber }}"></div>
                </div>
    ',
    link: (scope, elem, attrs) ->
        scope.replied = false
        scope.accepted = false
    , controller: ['$scope', '$location', 'api', ($scope, $location, api) ->
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
                        $scope.$parent.unreadCount -= 1
                    )
            )
            
        $scope.performAction = () ->
            actionCallback = () ->
                $scope.notification.read = true
    
                switch $scope.notification.notification_type
                    when 2, 3
                        $location.path('/u/' + $scope.notification.user_action.username.toLowerCase())
                    when 4, 5, 6
                        $location.path('/posts/show/' + $scope.notification.post_id)

            if not $scope.notification.read
                api(
                    '/api/notifications/seen',
                    'PUT',
                    notification_id: $scope.notification.id,
                    (resp) ->
                        $scope.$apply(actionCallback)
                )
            else
                actionCallback()
    ]
)