'use strict'

angular.module('sunglasses')
# displays a confirm dialog
.directive('confirmDialog', () ->
    restrict: 'E',
    replace: true,
    template: '<div class="ui modal small" id="confirm-dialog">
                  <i class="close icon"></i>
                  <div class="header" ng-bind="confirm.title"></div>
                  <div class="content">
                    <div class="left" ng-bind="confirm.message"></div>
                  </div>
                  <div class="actions">
                    <div class="ui button"  ng-class="confirm.dismissClass" ng-bind="confirm.cancel" ng-click="confirm.dismissCallback()"></div>
                    <div class="ui button" ng-class="confirm.acceptClass" ng-bind="confirm.accept" ng-click="confirm.acceptCallback()"></div>
                  </div>
                </div>'
)
.factory('confirm', ['$translate', ($translate) ->
    confirm =
        title: '',
        message: '',
        cancel: '',
        accept: '',
        acceptClass: 'negative',
        dismissClass: '',
        dismissCallback: () ->
            null
        , acceptCallback: () ->
            null
        , showDialog: (title, message, cancel, accept, acceptCallback, dismissCallback, acceptClass, dismissClass) ->
            $translate(title)
            .then((t) ->
                confirm.title = t
                $translate(message)
                .then((msg) ->
                    confirm.message = msg
                    $translate(cancel)
                    .then((_cancel) ->
                        confirm.cancel = _cancel
                        $translate(accept)
                        .then((_accept) ->
                            confirm.accept = _accept
                            confirm.acceptCallback = if acceptCallback? then acceptCallback else confirm.acceptCallback
                            confirm.dismissCallback = if dismissCallback? then dismissCallback else confirm.dismissCallback
                            confirm.acceptClass = if acceptClass? then acceptClass else confirm.acceptClass
                            confirm.dismissClass = if dismissClass? then dismissClass else confirm.dismissClass
                            $('#confirm-dialog').modal('show',
                                onApprove: confirm.acceptCallback,
                                onDeny: confirm.dismissCallback
                            )
                        )
                    )
                )
            )
            return
            
    return confirm
])