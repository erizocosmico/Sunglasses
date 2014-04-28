'use strict'

angular.module('mask.controllers')
.controller('HomeController', ['$scope', '$rootScope', ($scope, $rootScope) ->
   $scope.posts = []
   $scope.numPosts = 0
])
