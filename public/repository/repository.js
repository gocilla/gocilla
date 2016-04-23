'use strict';

angular.module('repository', ['ngResource'])

  .factory('RepositoryBuildsService', ['$resource', '$cacheFactory', function($resource, $cacheFactory) {
    var repositoryBuildsCache = $cacheFactory('repositoryBuildsCache');
    return $resource(
      '/api/organizations/:orgId/repositories/:repoId/builds',
      {orgId: '@orgId', repoId: '@repoId'},
      {query: {cache: repositoryBuildsCache, isArray: true}}
    );
  }])

  .controller('RepositoryController', [
        '$scope', '$routeParams', '$cacheFactory', '$http', 'RepositoryBuildsService',
        function($scope, $routeParams, $cacheFactory, $http, RepositoryBuildsService) {
    $scope.orgId = $routeParams.orgId;
    $scope.repoId = $routeParams.repoId;
    $scope.buildId = $routeParams.buildId;
    $scope.builds = RepositoryBuildsService.query({}, {orgId: $scope.orgId, repoId: $scope.repoId});
    if ($scope.buildId) {
      var buildLogsUrl = '/api/organizations/' + $scope.orgId + '/repositories/' + $scope.repoId + '/builds/'
          + $scope.buildId + '/logs';
      $http({method: 'GET', url: buildLogsUrl}).then(function(response) {
        $scope.buildLogs = response.data;
      });
    }
  }]);
