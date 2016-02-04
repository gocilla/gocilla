'use strict';

angular.module('triggers', ['ngResource'])

  .factory('TriggersService', ['$resource', '$cacheFactory', function($resource, $cacheFactory) {
    var triggersCache = $cacheFactory('triggersCache');
    return $resource('/api/triggers', {}, {query: {cache: triggersCache, isArray: true}});
  }])

  .controller('TriggersController', ['$scope', '$routeParams', '$cacheFactory',
        'TriggersService',
        function($scope, $routeParams, $cacheFactory, TriggersService) {
    $scope.orgId = $routeParams.orgId;
    $scope.repoId = $routeParams.repoId;
    $scope.triggerId = $routeParams.triggerId;
    $scope.triggers = TriggersService.query({organization: $scope.orgId, repository: $scope.repoId});
    $scope.trigger = {organization: $scope.orgId, repository: $scope.repoId, event: 'push', envVars: []};
    $scope.newEnvVar = {};

    $scope.addEnvVar = addEnvVar;
    $scope.deleteEnvVar = deleteEnvVar;
    $scope.submit = submit;

    if ($scope.triggerId) {
      $scope.triggers.$promise.then(function onTriggers() {
        try {
          var trigger = $scope.triggers[$scope.triggerId];
          if (trigger !== null) {
            $scope.trigger = trigger;
          }
        } catch(e) {
          console.log('Error getting trigger ' + $scope.triggerId)
        }
      });
    }

    function addEnvVar() {
      $scope.trigger.envVars.push($scope.newEnvVar);
      $scope.newEnvVar = {};
    }

    function deleteEnvVar(index) {
      $scope.trigger.envVars.splice(index, 1);
    }

    function submit() {
      var future = TriggersService.save($scope.trigger);
      future.$promise.then(function onSaved() {
        $cacheFactory.get('triggersCache').remove('/api/triggers?organization=' + $scope.orgId + '&repository=' + $scope.repoId);
        $scope.triggers = TriggersService.query({organization: $scope.orgId, repository: $scope.repoId});
      });
    }

  }]);
