'use strict'

angular.module('sunglasses.services')
# gives functionality to interact with user entities
.factory('user', ['$rootScope', 'api', ($rootScope, api) ->
    # retrieves the user avatar thumbnail (it can be public or private)
    user =
        getAvatarThumb: (user) ->
            if user.avatar_thumbnail?
                return user.avatar_thumbnail
            else
                return user.public_avatar_thumbnail
        # retrieves the user avatar (it can be public or private)   
        , getAvatar: (post) ->
            if user.avatar?
                return user.avatar
            else
                return user.public_avatar   
        # retrieves the username (it can be public or private)
        , getUsername: (user) ->
            if user.private_name? and user.private_name != ''
                return user.private_name
            else if user.public_name? and user.public_name != ''
                return user.public_name
            else
                return user.username
        , search: (query, justFollowings, offset, count, successCallback, errorCallback) ->
            data =
                q: query

            if justFollowings?
                data.justFollowings = justFollowings
            if offset?
                data.offset = offset
            if count?
                data.count = count

            api('/api/search'
                'GET',
                data,
                (resp) ->
                    successCallback(resp)
                , (resp) ->
                    errorCallback(resp.responseJSON)
            )
        # Send a follow request
        , sendFollowRequest: (user, isRequest) ->
            api(
                '/api/users/follow',
                'POST',
                user_to: user.id,
                (resp) ->
                    $rootScope.$apply(() ->
                        user.followed = true
                        user.follow_requested = !!isRequest
                    )
            )
        # Unfollow an user
        , unfollow: (user) ->
            api(
                '/api/users/unfollow',
                'DELETE',
                user_to: user.id,
                (resp) ->
                    $rootScope.$apply(() ->
                        user.followed = false
                    )
            )
                
    return user
])