'use strict'

angular.module('sunglasses')
.directive('postForm', () ->
    restrict: 'E',
    replace: true,
    templateUrl: 'templates/post-form.html',
    controller: ['$scope', 'api', '$rootScope', '$timeout', ($scope, api, $rootScope, $timeout) ->
        $scope.privacyOpened = false
        $scope.privacySelected = 4
        
        getUserPrivacy = (t) ->
            if t? and not userData.settings['override_default_privacy'] and ['status', 'photo', 'video', 'link'].indexOf(t) != -1
                return userData.settings['default_' + t + '_privacy'].privacy_type
                
            return userData.settings['default_status_privacy'].privacy_type
        
        $scope.privacySelect = (p) ->
            $scope.privacySelected = p
        
        # newPost creates a new empty post and changes the post status
        # that means it initializes the post-box to send another post after
        # submitting a post
        newPost = () ->
            if 'changePostType' in $scope then $scope.changePostType('status')
            document.getElementById('photo-upload').value = ''
            
            $scope.privacySelected = getUserPrivacy()
            
            text: '',
            url: '',
            type: 'status',
            caption: '',
            picture: null

        # post contains the data used to create new posts
        $scope.post = newPost()

        document.getElementById('photo-upload').addEventListener('change', (e) ->
            $scope.post.picture = e.target.files[0]
            document.getElementById('filename').innerHTML = $scope.post.picture.name
        )
        
        # submits a post to the server
        $scope.submitPost = () ->
            $scope.privacyOpened = false
            urlRegex = /^https?:\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?$/
            vimeoReg = /^https?:\/\/(www.)?vimeo.com\/([0-9]+)$/
            ytReg = /^https?:\/\/(www.)?youtube.com\/watch?(.*)v=(.+)$/

            if $scope.post.text.trim().length == 0 && $scope.post.type == 'status'
                $rootScope.showMsg('error_post_text_empty', 'post-error')
            else if $scope.post.text.length > 1500
                $rootScope.showMsg('error_post_text_too_long', 'post-error')
            else if $scope.post.type == 'link' and !urlRegex.test($scope.post.url)
                $rootScope.showMsg('error_post_invalid_url', 'post-error')
            else if $scope.post.type == 'video' and !(vimeoReg.test($scope.post.url) or ytReg.test($scope.post.url))
                $rootScope.showMsg('error_post_invalid_video_url', 'post-error')
            else   
                api(
                    '/api/posts/create',
                    'POST',
                    post_type: $scope.post.type,
                    post_text: $scope.post.text,
                    post_url: $scope.post.url,
                    post_picture: $scope.post.picture,
                    caption: $scope.post.caption,
                    privacy_type: $scope.privacySelected,
                    (resp) ->
                        $scope.loading = true
                        $scope.post = newPost()
                        $rootScope.showAlert('post_success', true)
                        #$rootScope.showMsg('post_success', 'post-success', true)
                        window.setTimeout(() ->
                            $scope.loadPosts('newer')
                        , 4000)
                )
                
        $scope.handleUpload = () ->
            e = document.createEvent('Event')
            e.initEvent('click', true, true)
            document.getElementById('photo-upload').dispatchEvent(e)
            
        # changes the post type
        $scope.changePostType = (type) ->
            if ['status', 'photo', 'video', 'link'].indexOf(type) == -1 then type = 'status'
            $scope.privacySelected = getUserPrivacy(type)
            if $scope.post.type != type then $scope.post.url = ''
            $scope.post.type = type
            btns = document.querySelectorAll('.post-box ul li span')
            for btn in btns
                if btn.className.indexOf('active') != -1
                    btn.className = btn.className.replace('active', '')
            document.getElementById('post-type-' + type).className += ' active'
            $scope.post.url = ''
    ]
)