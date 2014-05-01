'use strict'

angular.module('mask.services', ['ngRoute'])
angular.module('mask.controllers', ['mask.services'])
angular.module('mask', ['ngRoute', 'ngCookies', 'mask.controllers', 'mask.services', 'pascalprecht.translate'])
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
        $translateProvider.preferredLanguage('en')
        $translateProvider.useSanitizeValueStrategy('escaped')
        $translateProvider.useLocalStorage()
        $translateProvider.fallbackLanguage('en')

        if userData?
            $routeProvider
            .when('/profile',
                templateUrl: 'templates/profile.html'
                controller: 'ProfileController'
            )
            .when('/',
                templateUrl: 'templates/home.html'
                controller: 'HomeController'
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
                templateUrl: 'templates/landing.html'
            )

        $routeProvider.otherwise(
            redirectTo: '/'
        )
])
.run(['$rootScope', '$translate', '$location', ($rootScope, $translate, $location) ->
    document.getElementsByTagName('title')[0].innerHTML = '{{ title | translate }}'
    $rootScope.title = 'Mask'
    
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
            callback(elem)
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
])