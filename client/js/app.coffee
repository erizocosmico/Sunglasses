'use strict'

angular.module('mask.controllers', [])
angular.module('mask.services', ['ngRoute'])
angular.module('mask', ['ngRoute', 'mask.controllers', 'mask.services'])
.config(['$routeProvider', '$locationProvider', ($routeProvider, $locationProvider) ->
    $locationProvider.html5Mode(false)

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