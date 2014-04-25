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
        
        #languagesProvider.registerLanguages($translateProvider)
        $translateProvider.useStaticFilesLoader(
            prefix: 'lang/',
            suffix: '.json'
        )
        $translateProvider.preferredLanguage('es')
        $translateProvider.useSanitizeValueStrategy('escaped')
        $translateProvider.useLocalStorage()

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

angular.module('mask').value('bullshit', 'caca')