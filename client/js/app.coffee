'use strict'

mask = angular.module('mask', [
    'ngRoute'
])

mask.config(['$routeProvider', '$rootScope'], ($routeProvider, $rootScope) ->
    $locationProvider.html5Mode(false)

    #userData is rendered server-side when accessing /
    $rootScope.loggedIn = userData?
    $rootScope.refresh = () ->
        window.location.href = window.location.href

    if $rootScope.loggedIn
        $rootScope.user = userData

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
)
