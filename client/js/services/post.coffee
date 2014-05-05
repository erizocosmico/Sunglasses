'use strict'

angular.module('sunglasses.services')
# gives functionality to interact with post entities
.factory('post', () ->
    # tells if the post is a certain kind of post
    isKindOf: (post, kind) ->
        switch kind
            when 'video'
                return post.video_id? and post.video_service?
            when 'link'
                return post.link_url?
            when 'photo'
                return post.photo_url? and post.thumbnail?
        return true
    # returns the post title or its url in case it does not have a title
    , getLinkTitle: (post) ->
        return if post.title? then post.title else post.link_url
)