'use strict';

angular.module('gocilla', ['ngResource', 'ngRoute', 'builds', 'profile', 'repositories', 'triggers'])

  .config(function ($routeProvider, $locationProvider) {
    $routeProvider
      .when('/', {
        templateUrl: 'repositories/repositories.html',
        controller: 'RepositoriesController'
      })
      .when('/builds', {
        templateUrl: 'builds/builds.html',
        controller: 'BuildsController'
      })
      .when('/builds/:buildId', {
        templateUrl: 'builds/builds.html',
        controller: 'BuildsController'
      })
      .when('/organizations', {
        templateUrl: 'repositories/repositories.html',
        controller: 'RepositoriesController'
      })
      .when('/organizations/:orgId', {
        templateUrl: 'repositories/repositories.html',
        controller: 'RepositoriesController'
      })
      .when('/organizations/:orgId/repositories/:repoId/hook', {
        templateUrl: 'repositories/repositories.html',
        controller: 'RepositoriesController'
      })
      .when('/organizations/:orgId/repositories/:repoId/triggers', {
        templateUrl: 'triggers/triggers.html',
        controller: 'TriggersController'
      })
      .when('/organizations/:orgId/repositories/:repoId/triggers/:triggerId', {
        templateUrl: 'triggers/triggers.html',
        controller: 'TriggersController'
      })
      .otherwise({
        redirectTo: '/'
      });
    $locationProvider.html5Mode(true);
  })
  .filter('moment', function () {
    return function (input, momentFn /*, param1, param2, ...param n */) {
      var args = Array.prototype.slice.call(arguments, 2),
          momentObj = moment(input);
      return momentObj[momentFn].apply(momentObj, args);
    };
  })
  .filter('duration', function () {
    return function (start, end) {
      var m1 = moment(start);
      var m2 = moment(end);
      return moment.duration(m1.diff(m2)).humanize();
    };
  })
  .directive('navigation', function() {
    return {
      restrict: 'AE',
      replace: 'true',
      templateUrl: 'templates/navigation.html'
    }
  });
