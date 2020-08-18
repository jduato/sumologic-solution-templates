#Sumo Logic SDO solution - This files has parameters required to create FER's in Sumo Logic.

github_pull_request_fer_scope = "%\"x-github-event\"=pull_request"
github_pull_request_fer_parse = "json \"action\", \"pull_request.title\", \"pull_request.created_at\", \"pull_request.merged_at\" ,\"repository.name\" ,\"pull_request.merged\", \"pull_request.html_url\", \"pull_request.merge_commit_sha\", \"pull_request.base.ref\" , \"pull_request.user.login\", \"pull_request.labels[0].name\", \"pull_request.requested_reviewers[*].login\" as action, title, dateTime, closeddate ,repository_name,  merge, link, commit_id, target_branch ,user, service, reviewers nodrop\n           | where action in (\"closed\", \"opened\") \n | parseDate(dateTime, \"yyyy-MM-dd'T'HH:mm:ss\") as dateTime_epoch\n | if(action matches \"closed\" and merge matches \"true\", \"merged\", if(action matches \"closed\" and merge matches \"false\", \"declined\", if (action matches \"opened\", \"created\", \"other\"  ))) as status\n   | if (status=\"merged\", parseDate(closeddate, \"yyyy-MM-dd'T'HH:mm:ss\") , 000000000 ) as closeddate_epoch\n | toLong(closeddate_epoch)\n | \"pull_request\" as event_type\n"

jenkins_build_fer_scope = "Job_Status"
jenkins_build_fer_parse = "json \"result\", \"description\", \"name\", \"user\" ,\"label\", \"jobBuildURL\", \"start_time\" as status, message, title, user, target_branch, link, datetime_epoch\n            | toLong(datetime_epoch) as datetime_epoch\n     // setting repository_name as empty\n            | \"\" as repository_name\n        | if ( status matches \"SUCCESS\", \"SUCCESSFUL\", if(status matches \"FAILURE\", \"FAILED\", status)) as status\n            | where status in (\"SUCCESSFUL\", \"FAILED\")\n            | \"build\" as event_type\n"

jenkins_deploy_fer_scope = "timestamp commit_id result event_name env_name git_url"
jenkins_deploy_fer_parse = "json field=_raw \"event_name\", \"result\", \"commit_id\", \"env_name\", \"git_url\", \"timeStamp\" as event_type, status, commit_id, environment_name, link, datetime\n    | toLong(datetime) as datetime_epoch\n    | parse regex field=link \"(?<repository_name>[^/]+)\\.git$\"\n    | parse regex field=link \"(?<link>.+)\\.git\"\n    | concat (link,\"/commit/\", commit_id) as link\n    | if (status matches \"SUCCESS\", \"Success\", if (status matches \"FAILURE\", \"Failed\", status) ) as status\n    | \"deploy\" as event_type\n"

opsgenie_alerts_fer_scope = "(\"Close\" or \"Create\")"
opsgenie_alerts_fer_parse = "json \"alert.createdAt\", \"alert.updatedAt\", \"action\", \"alert.team\",  \"alert.priority\", \"alert.source\", \"alert.alertId\" as dateTime, closeddate, alert_type, team, priority, service, alert_id nodrop\n| where alert_type in (\"Close\", \"Create\") \n| toLong(closeddate/1000) as closeddate_epoch\n| toLong(datetime*1000) as datetime_epoch\n| if (priority matches \"p1\", \"high\", if(priority matches \"p2\", \"medium\",  if(priority matches \"p3\", \"medium\",  if(priority matches \"p4\", \"low\",  if(priority matches \"p5\", \"low\", \"other\")))))  as priority\n| if (alert_type matches \"*Create\", \"alert_created\", if(alert_type matches \"*Close\", \"alert_closed\", \"other\") ) as event_type\n"

bitbucket_pull_request_fer_scope = "%\"x-event-key\"=pullrequest:*"
bitbucket_pull_request_fer_parse = "json \"pullrequest.title\",  \"pullrequest.destination.repository.full_name\", \"pullrequest.destination.branch.name\", \"pullrequest.created_on\", \"pullrequest.author.display_name\",  \"pullrequest.state\",  \"pullrequest.links.html.href\", \"pullrequest.merge_commit.hash\", \"pullrequest.reviewers[*].display_name\", \"pullrequest.updated_on\" as  title, repository_name, target_branch, dateTime, user, action, link, commit_id, reviewers, closeddate nodrop\n            | parse regex field= datetime \"(?<datetime_epoch>\\d\\d\\d\\d-\\d\\d-\\d\\dT\\d\\d:\\d\\d:\\d\\d)\"  nodrop   \n| parse regex field= closeddate \"(?<closeddate_epoch>\\d\\d\\d\\d-\\d\\d-\\d\\dT\\d\\d:\\d\\d:\\d\\d)\"  nodrop   \n            | parseDate(datetime_epoch, \"yyyy-MM-dd'T'HH:mm:ss\") as datetime_epoch\n            | parseDate(closeddate_epoch, \"yyyy-MM-dd'T'HH:mm:ss\") as closeddate_epoch\n            | parse regex field=%\"x-event-key\" \".+:(?<status>.+)\"\n            | if (status matches \"created\", \"created\", if(status matches \"fulfilled\", \"merged\", if(status matches \"rejected\", \"declined\", \"other\"))) as status \n            | \"pull_request\" as event_type\n"

bitbucket_build_fer_scope = "%\"x-event-key\"=repo:commit_status_* (\"SUCCESSFUL\" OR \"FAILED\")"
bitbucket_build_fer_parse = "json \"commit_status.state\", \"commit_status.commit.message\", \"commit_status.name\", \"actor.display_name\", \"repository.full_name\",  \"commit_status.refname\", \"commit_status.url\", \"commit_status.created_on\" as status, message, title, user, repository_name, target_branch, link, dateTime\n            | parse regex field= datetime \"(?<datetime_epoch>\\d\\d\\d\\d-\\d\\d-\\d\\dT\\d\\d:\\d\\d:\\d\\d)\"  nodrop   \n            | parseDate(datetime_epoch, \"yyyy-MM-dd'T'HH:mm:ss\") as datetime_epoch\n            | where status in (\"SUCCESSFUL\", \"FAILED\")\n            | \"build\" as event_type\n"

bitbucket_deploy_fer_scope = "deploymentEnvironment pipe_result_link deploy_status commit_link"
bitbucket_deploy_fer_parse = "json field=_raw  \"deploymentEnvironment\",  \"repoFullName\", \"pipe_result_link\", \"deploy_status\", \"commit\",  \"event_date\"  as environment_name,  repository_name, link, status, commit_id, dateTime\n| parse regex field=dateTime \"(?<datetime_epoch>[\\S]+) (?<dateTime>[\\S]+)\"\n| concat(datetime_epoch,\"T\",dateTime) as datetime_epoch\n| parseDate(datetime_epoch, \"yyyy-MM-dd'T'HH:mm:ss\") as datetime_epoch\n| if (status matches \"0\", \"Success\", \"Failed\") as status\n| \"deploy\" as event_type\n"

jira_issues_fer_scope = "issue_*"
jira_issues_fer_parse = "json field=_raw \"issue_event_type_name\", \"issue.self\",  \"issue.key\", \"issue.fields.issuetype.name\", \"issue.fields.project.name\", \"issue.fields.status.statusCategory.name\",  \"issue.fields.priority.name\" as  issue_event_type, link, issue_key, issue_type, project_name, issue_status,  priority \n| json \"issue.fields.resolutiondate\", \"issue.fields.created\" as closedDate, dateTime \n| parseDate(dateTime, \"yyyy-MM-dd'T'HH:mm:ss.SSSZ\") as datetime_epoch\n| if (isNull(closeddate) , 00000000000, parseDate(closedDate, \"yyyy-MM-dd'T'HH:mm:ss.SSSZ\") ) as closeddate_epoch\n| toLong(closeddate_epoch)\n| \"issue\" as event_type\n"

pagerduty_alerts_fer_scope = "(\"incident.trigger\" or \"incident.resolve\" )"
pagerduty_alerts_fer_parse = "parse regex \"(?<event>\\{\\\"event\\\":\\\"incident\\..+?\\}(?=,\\{\\\"event\\\":\\\"incident\\..+|\\]\\}$))\" \n|json  field=event \"event\", \"created_on\", \"incident\" as alert_type, dateTime, incident\n|json field=incident \"id\",  \"service.name\" , \"urgency\", \"teams[0].summary\", \"html_url\"  as alert_id, service, priority, team, link\n|json  field=incident \"created_at\" as closeddate nodrop\n| where alert_type in (\"incident.trigger\", \"incident.resolve\")\n| parseDate(dateTime, \"yyyy-MM-dd'T'HH:mm:ss'Z'\") as dateTime_epoch\n| parseDate(closeddate, \"yyyy-MM-dd'T'HH:mm:ss'Z'\") as closeddate_epoch\n| if (alert_type matches \"*trigger\", \"alert_created\", if(alert_type matches \"*resolve\", \"alert_closed\", \"other\") ) as event_type\n"