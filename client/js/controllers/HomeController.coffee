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
    ($scope, $rootScope, $translate, api, user, post, photo) ->
        # number of posts retrieved
        $scope.postCount = 0
        # array of the previously retrieved posts
        $scope.posts = []
        # loading shows the loading dialog when new posts are being loaded
        $scope.loading = true
        $scope.userService = user
        $scope.postService = post
        $scope.photoService = photo
        
        # newPost creates a new empty post and changes the post status
        # that means it initializes the post-box to send another post after
        # submitting a post
        newPost = () ->
            if 'changePostType' in $scope then $scope.changePostType('status')
            document.getElementById('photo-upload').value = ''
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
            
        # shows an error or a success message
        showMsg = (text, field, success) ->
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
                    showMsg('error_code_' + resp.responseJSON.code, 'timeline-error')
            )
            
        # submits a post to the server
        # TODO: privacy handling
        $scope.submitPost = () ->
            urlRegex = /^https?:\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?$/
            vimeoReg = /^https?:\/\/(www.)?vimeo.com\/([0-9]+)$/
            ytReg = /^https?:\/\/(www.)?youtube.com\/watch?(.*)v=(.+)$/

            if $scope.post.text.trim().length == 0 && $scope.post.type == 'status'
                showMsg('error_post_text_empty', 'post-error')
            else if $scope.post.text.length > 1500
                showMsg('error_post_text_too_long', 'post-error')
            else if $scope.post.type == 'link' and !urlRegex.test($scope.post.url)
                showMsg('error_post_invalid_url', 'post-error')
            else if $scope.post.type == 'video' and !(vimeoReg.test($scope.post.url) or ytReg.test($scope.post.url))
                showMsg('error_post_invalid_video_url', 'post-error')
            else   
                api(
                    '/api/posts/create',
                    'POST',
                    post_type: $scope.post.type,
                    post_text: $scope.post.text,
                    post_url: $scope.post.url,
                    post_picture: $scope.post.picture,
                    caption: $scope.post.caption,
                    (resp) ->
                        $scope.loading = true
                        $scope.post = newPost()
                        showMsg('post_success', 'post-success', true)
                        window.setTimeout(() ->
                            $scope.loadPosts(true)
                        , 4000)
                    , (resp) ->
                        showMsg('error_code_' + resp.responseJSON.code, 'post-error')
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
                    showMsg('error_code_' + resp.responseJSON.code, 'post-error')
            )
                
        $scope.handleUpload = () ->
            e = document.createEvent('Event')
            e.initEvent('click', true, true)
            document.getElementById('photo-upload').dispatchEvent(e)
            
        # changes the post type
        $scope.changePostType = (type) ->
            if ['status', 'photo', 'video', 'link'].indexOf(type) == -1 then type = 'status'
            if $scope.post.type != type then $scope.post.url = ''
            $scope.post.type = type
            btns = document.querySelectorAll('.post-box ul li span')
            for btn in btns
                if btn.className.indexOf('active') != -1
                    btn.className = btn.className.replace('active', '')
            document.getElementById('post-type-' + type).className += ' active'
            $scope.post.url = ''
                
        $scope.loadPosts()
])
