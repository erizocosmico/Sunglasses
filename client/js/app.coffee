'use strict'

angular.module('sunglasses.services', ['ngRoute'])
angular.module('sunglasses.controllers', ['sunglasses.services'])
angular.module('sunglasses', ['ngRoute', 'ngCookies', 'sunglasses.controllers', 'sunglasses.services', 'pascalprecht.translate'])
.config([
    '$routeProvider',
    '$locationProvider',
    '$translateProvider',
    ($routeProvider, $locationProvider, $translateProvider) ->
        $locationProvider.html5Mode(false)

        $translateProvider.useStaticFilesLoader(
            prefix: 'lang/',
            suffix: '.json'
        )
        $translateProvider.preferredLanguage(userData?.preferred_lang || 'en')
        $translateProvider.useSanitizeValueStrategy('escaped')
        $translateProvider.useLocalStorage()
        $translateProvider.fallbackLanguage('en')

        if userData?
            $routeProvider
            .when('/u/:username',
                templateUrl: 'templates/profile.html'
                controller: 'ProfileController'
            )
            .when('/posts/show/:postid',
                templateUrl: 'templates/home.html'
            )
            .when('/settings',
                templateUrl: 'templates/settings.html'
                controller: 'SettingsController'
            )
            .when('/',
                templateUrl: 'templates/home.html'
            )
            .otherwise(
                redirectTo: '/'
            )
        else
            $routeProvider
            .when('/login',
                templateUrl: 'templates/login.html'
                controller: 'LoginController'
            )
            .when('/signup',
                templateUrl: 'templates/signup.html'
                controller: 'SignupController'
            )
            .when('/recover',
                templateUrl: 'templates/recover.html'
                controller: 'RecoverController'
            )
            .when('/',
                controller: 'LandingController'
                templateUrl: 'templates/landing.html'
            )

        $routeProvider.otherwise(
            redirectTo: '/'
        )
])
.run(['$rootScope', '$translate', '$location', ($rootScope, $translate, $location) ->
    document.getElementsByTagName('title')[0].innerHTML = '{{ title | translate }}'
    $rootScope.title = 'sunglasses'
    
    $rootScope.userData = userData
    
    # changes the language of the application
    $rootScope.changeLang = (lang) ->
        $translate.use(lang);
        
    # redirects the user to the home
    $rootScope.goHome = () ->
        $location.path('/')
        
    $rootScope.redirect = (path) ->
        $location.path(path)
        
    # refreshes the page so that the frontend is loaded again (logged in or logged out)
    $rootScope.fullRefresh = () ->
        window.location.href = window.location.href
            .substring(0, window.location.href.indexOf('#'))
        
    # animate element using animate.css and perform a callback on completion
    $rootScope.animateElem = (elem, animation, callback) ->
        elem.className = 'animated ' + animation
        $(elem).one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', () ->
            if callback? then callback(elem)
        )
        
    lastField = null
    lastTimeout = null
        
    # display error or success
    $rootScope.displayError = (field, success) ->
        classType = if success? then 'success' else 'error'
        elem = document.getElementById(field)
        elem.className = classType + ' animated fadeInUp'
        lastField = field
        lastTimeout = window.setTimeout(() ->
            if lastField == field
                window.clearTimeout(lastTimeout)
            if elem.className.indexOf('hidden') == -1
                $rootScope.animateElem(
                    elem,
                    'fadeOutDown',
                    (el) -> el.className = classType + ' hidden'
                )
        , 6000)
        
    # shows an error or a success message
    $rootScope.showMsg = (text, field, success) ->
        $translate(text).then (msg) ->
            document.getElementById(field).innerHTML = msg
            $rootScope.displayError(field, success)
    
    # shows the alert box
    $rootScope.showAlert = (text, translate, danger) ->
        displayAlert = (msg, danger) ->
            alertBox = document.getElementById('alert-box')
            alertBox.getElementsByTagName('p')[0].innerHTML = msg

            alertBox.className = 'alert-box'
            if danger?
                alertBox.className += ' danger'
            
            window.setTimeout(() ->
                alertBox.className += ' hidden'
            , 3000)
        
        if translate?
            $translate(text).then (msg) ->
                console.log(msg)
                displayAlert(msg, danger)
        else
            displayAlert(text, danger)
        
    # returns the relative time
    $rootScope.relativeTime = (time, dict, callback) ->
        diff = Math.floor(new Date().getTime() / 1000) - Number(time)
        unit = 's'
        num = diff

        # Minutes
        if diff < 3600
            num = Math.ceil(diff / 60)
            unit = 'm'
        # Hours
        else if diff < 86400
            num = Math.floor(diff / 3600)
            unit = 'h'
        # Days
        else
            num = Math.floor(diff / 86400)
            unit = 'd'
                
        $translate(unit)
        .then((s) ->
            $translate('time_format')
            .then((t) ->
                str = t.replace('%num%', num).replace('%unit%', s)
                if dict?
                    dict.timeFormatted = str
                else
                    callback(str)
            )
        )
        
    $rootScope.displayIntro = false
    if window.localStorage?
        $rootScope.displayIntro = window.localStorage.getItem('just_signed_up')
])
.directive('compile', ['$compile', ($compile) ->
    (scope, element, attrs) ->
        scope.$watch(
            (scope) ->
                scope.$eval(attrs.compile);
            , (value) ->
                element.html(value);
                $compile(element.contents())(scope);
        )
])