'use strict'

angular.module('sunglasses')
.directive('intro', () ->
    restrict: 'E',
    replace: true,
    templateUrl: 'templates/intro.html',
    controller: ['$scope', '$rootScope', 'api', '$timeout', ($scope, $rootScope, api, $timeout) ->
        dismiss = () ->
            intro = document.getElementById('intro')
            overlay = document.getElementsByClassName('intro-overlay')[0]
            intro.className += ' animated bounceOutUp'
            overlay.className += ' animated fadeOutUp'
            
            $timeout(() ->
                $rootScope.displayIntro = false
            , 500)
            
            if window.localStorage?
                window.localStorage.setItem('just_signed_up', false)
        
        $scope.ok = (type) ->
            switch type
                when 1
                    dismiss()
                when 2, 3
                    params =
                        invisible: 'false',
                        can_receive_requests: 'true',
                        notify_new_comment: 'true',
                        notify_new_comment_others: 'true',
                        notify_posts_in_my_profile: 'true',
                        notify_likes: 'true',
                        allow_posts_in_my_profile: 'true',
                        allow_comments_in_posts: 'true',
                        override_default_privacy: 'true'
                    
                    if type == 2
                        params.follow_approval_required = 'true'
                        params.default_status_privacy = 2
                        params.privacy_status_type = 2
                    else
                        params.follow_approval_required = 'false'
                        params.default_status_privacy = 1
                        params.privacy_status_type = 1
                        params.display_avatar_before_approval = 'true'
                
                    api(
                        '/api/account/settings',
                        'PUT',
                        params,
                        (resp) ->
                            dismiss()
                            $rootScope.$apply(() ->
                                for k, v of params
                                    if v == 'true' then v = true
                                    if v == 'false' then v = false
                                    $rootScope.userData.settings[k] = v
                            )
                    ) 
    ]
)