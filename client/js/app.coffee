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
                controller: 'LandingController'
            )

        $routeProvider.otherwise(
            redirectTo: '/'
        )
])
.run(['$rootScope', '$translate', '$location', ($rootScope, $translate, $location) ->
    document.getElementsByTagName('title')[0].innerHTML = '{{ title | translate }}'
    $rootScope.title = 'Mask'
    
    $rootScope.changeLang = (lang) ->
        $translate.use(lang);
        
    $rootScope.goHome = () ->
        $location.path('/')
        
    $rootScope.animateElem = (elem, animation, callback) ->
        elem.className = 'animated ' + animation
        $(elem).one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', () ->
            callback(elem)
        )
])