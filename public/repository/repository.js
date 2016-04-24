'use strict';

angular.module('repository', ['ngResource'])
  .factory('RepositoryService', RepositoryService)
  .factory('RepositoryBuildsService', RepositoryBuildsService)
  .controller('RepositoryController', RepositoryController)
  .controller('RepositoryBuildController', RepositoryBuildController)
  .controller('RepositorySettingsController', RepositorySettingsController);

RepositoryService.$inject = ['$resource', '$cacheFactory'];
function RepositoryService($resource, $cacheFactory) {
  var repositoryCache = $cacheFactory('repositoryCache');
  return $resource(
    '/api/organizations/:orgId/repositories/:repoId',
    {orgId: '@orgId', repoId: '@repoId'},
    {get: {cache: repositoryCache}}
  );
}

RepositoryBuildsService.$inject = ['$resource', '$cacheFactory'];
function RepositoryBuildsService($resource, $cacheFactory) {
  var repositoryBuildsCache = $cacheFactory('repositoryBuildsCache');
  return $resource(
    '/api/organizations/:orgId/repositories/:repoId/builds',
    {orgId: '@orgId', repoId: '@repoId'},
    {query: {cache: repositoryBuildsCache, isArray: true}}
  );
}

RepositoryController.$inject = ['$scope', '$routeParams', '$cacheFactory', 'RepositoryBuildsService'];
function RepositoryController($scope, $routeParams, $cacheFactory, RepositoryBuildsService) {
  $scope.orgId = $routeParams.orgId;
  $scope.repoId = $routeParams.repoId;
  $scope.builds = RepositoryBuildsService.query({}, {orgId: $scope.orgId, repoId: $scope.repoId});
}

RepositoryBuildController.$inject = ['$scope', '$routeParams', '$cacheFactory', '$http', 'RepositoryBuildsService'];
function RepositoryBuildController($scope, $routeParams, $cacheFactory, $http, RepositoryBuildsService) {

  RepositoryController($scope, $routeParams, $cacheFactory, RepositoryBuildsService);

  $scope.buildId = $routeParams.buildId;

  updateLogs();

  function updateLogs() {
    var buildLogsUrl = '/api/organizations/' + $scope.orgId + '/repositories/' + $scope.repoId + '/builds/'
        + $scope.buildId + '/logs';
    $http({method: 'GET', url: buildLogsUrl}).then(function(response) {
      $scope.buildLogs = response.data;
    });
  }
}

RepositorySettingsController.$inject = ['$scope', '$routeParams', '$cacheFactory', 'RepositoryService', 'RepositoryBuildsService'];
function RepositorySettingsController($scope, $routeParams, $cacheFactory, RepositoryService, RepositoryBuildsService) {

  RepositoryController($scope, $routeParams, $cacheFactory, RepositoryBuildsService);

  $scope.addEnvVar = addEnvVar;
  $scope.deleteEnvVar = deleteEnvVar;
  $scope.submit = submit;
  $scope.cancel = cancel;

  $scope.repository = getRepository();
  $scope.newEnvVar = {};

  function addEnvVar() {
    if ($scope.repository.envVars === null) {
      $scope.repository.envVars = [];
    }
    $scope.repository.envVars.push($scope.newEnvVar);
    $scope.newEnvVar = {};
  }

  function deleteEnvVar(index) {
    $scope.repository.envVars.splice(index, 1);
  }

  function getRepository() {
    return RepositoryService.get({}, {orgId: $scope.orgId, repoId: $scope.repoId});
  }

  function submit() {
    RepositoryService.save($scope.repository).$promise.then(function onSaved() {
      var repositoryUrl = '/api/organizations/' + $scope.orgId + '/repositories/' + $scope.repoId;
      $cacheFactory.get('repositoryCache').remove(repositoryUrl);
      $scope.repository = getRepository();
    }, function onError() {
      console.log('Error');
    });
  }

  function cancel() {
    $scope.repository = getRepository();
  }
}
