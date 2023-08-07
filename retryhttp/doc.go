/*
Package retryhttp provides simple tasks around HTTP client operations
that work with retry.Runner and retry.RunnerWithData.

Task is the central type in this package.  It allows definition of the
HTTP client, how requests get created for each attempt, and how responses
are converted into results.
*/
package retryhttp
