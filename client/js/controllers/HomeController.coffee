'use strict'

angular.module('mask.controllers')
.controller('HomeController', ['$scope', '$rootScope', '$http', ($scope, $rootScope, $http) ->
   $scope.posts = []
   $scope.numPosts = 0

   # Get initial posts
   $http.get('api/timeline')
   .success((data) ->
      $scope.posts = data.posts
      $scope.numPosts = data.count

      if numPosts == 0
         # TODO: Show a message saying that there are no posts yet
         console.log('TODO')
   )
   .error((data) ->
      console.log('TODO')
   )

   # Load more posts
   $scope.loadMore = () ->
      $http(
         url: 'api/timeline'
         method: "GET"
         params:
            count: 25
            offset: $scope.numPosts
      )
      .success((data) ->
         $scope.posts.concat(data.posts)
         $scope.numPosts += data.count
      )
      .error((data) ->
         console.log('TODO')
      )
])
