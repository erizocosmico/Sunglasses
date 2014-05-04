'use strict'

angular.module('sunglasses.controllers')
.controller('SignupController', [
    '$scope',
    '$rootScope',
    'api',
    ($scope, $rootScope, api) ->
        # errorVisible determines if the error is visible or not
        $scope.errorVisible = false
        
        # user data
        $scope.user =
            username: '',
            password: '',
            password_repeat: '',
            recovery_method: 0,
            recovery_answer: '',
            recovery_question: ''
        
        # current section
        $scope.sections = [true, false, false, false]
        currentSection = 0
        submitted = false
        
        # display dropdowns
        $('.ui.dropdown').dropdown(
            onChange: (val) ->
                $('.error').addClass('hidden')
                $scope.user.recovery_method = parseInt(val)
                question = document.getElementById('recovery-question')
                answer = document.getElementById('recovery-answer')
                email = document.getElementById('recovery-email')
                info = document.getElementById('recovery-email-info')
                question.className = 'hidden'
                answer.className = 'hidden'
                email.className = 'hidden'
                info.className = 'hidden'

                if val == 1
                    email.className = 'inputbox'
                    info.className = ''
                else if val == 2
                    question.className = 'inputbox'
                    answer.className = 'inputbox'
        )
        
        # validators for all steps
        validators = [
            # validator for username step
            (callback) ->
                if /^[a-zA-Z_0-9]{2,30}$/.test($scope.user.username)
                    api('/api/account/username_taken',
                        'GET',
                        username: $scope.user.username,
                        (resp) ->
                            if !resp.taken
                                callback()
                            else
                                $rootScope.displayError('username-error')
                        , (resp) ->
                            $rootScope.displayError('username-error')
                    )
                else
                    $rootScope.displayError('username-error')
            # validator for password step
            , (callback) ->
                valid = true
                if $scope.user.password.length < 6
                    valid = false
                    $rootScope.displayError('password-error')
                    
                if $scope.user.password != $scope.user.password_repeat
                    valid = false
                    $rootScope.displayError('password-repeat-error')

                if valid then callback()
            , (callback) ->
                switch $scope.user.recovery_method
                    when 1
                        if not /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/.test($scope.user.recovery_email)
                            $rootScope.displayError('recovery-email-error')
                        else
                            callback()
                    when 2
                        valid = true
                        if $scope.user.recovery_question.length < 1
                            valid = false
                            $rootScope.displayError('recovery-question-error')
                    
                        if $scope.user.recovery_answer.length < 1
                            valid = false
                            $rootScope.displayError('recovery-answer-error')

                        if valid then callback()
                    else
                        callback()
        ]
        
        $rootScope.title = 'signup'
        
        # loginClick performs an api call to login the user
        # if successful the user will be redirected to the home page
        $scope.signupClick = () ->
            if password? or username?
                $scope.errorVisible = true
            else
                api('/api/account/signup',
                    'POST',
                    username: $scope.username
                    password: $scope.password,
                    (resp) ->
                        console.log ":)"
                    (resp) ->
                        $scope.errorVisible = true
                )
              
        # nextSection goes to next section if the fields of the current section are valid
        $scope.nextSection = () ->
            if currentSection < 3
                callback = () ->
                    circles = document.querySelectorAll('ul.step li')
                    sections = document.querySelectorAll('div.sections article')
                    descs = document.querySelectorAll('div.description p')
                    circles[currentSection].className = 'passed'
                    $rootScope.animateElem(sections[currentSection], 'bounceOutLeft', (el) -> 
                        el.className = 'hidden'
                        sections[currentSection].className = 'animated bounceInRight'
                    )
                    $rootScope.animateElem(descs[currentSection], 'bounceOutLeft', (el) ->
                        el.className = 'hidden'
                        descs[currentSection].className = 'animated bounceInRight'
                    )
                    currentSection += 1
                    circles[currentSection].className = 'active'
                        
                if currentSection+1 == 3
                    validators[currentSection](() ->
                        if not submitted
                            api(
                                '/api/account/signup',
                                'POST',
                                $scope.user,
                                (resp) ->
                                    callback()
                                , (resp) ->
                                    console.log ":("    
                            )
                            submitted = true
                    )
                else
                    validators[currentSection](callback)
            else
                $rootScope.fullRefresh()
])
