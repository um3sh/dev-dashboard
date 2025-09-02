export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace types {
	
	export class Action {
	    id: number;
	    repository_id: number;
	    service_id?: number;
	    resource_id?: number;
	    type: string;
	    status: string;
	    workflow_run_id: number;
	    commit: string;
	    branch: string;
	    build_hash: string;
	    started_at: time.Time;
	    completed_at?: time.Time;
	    created_at: time.Time;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Action(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repository_id = source["repository_id"];
	        this.service_id = source["service_id"];
	        this.resource_id = source["resource_id"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.workflow_run_id = source["workflow_run_id"];
	        this.commit = source["commit"];
	        this.branch = source["branch"];
	        this.build_hash = source["build_hash"];
	        this.started_at = this.convertValues(source["started_at"], time.Time);
	        this.completed_at = this.convertValues(source["completed_at"], time.Time);
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ActionWithDetails {
	    id: number;
	    repository_id: number;
	    service_id?: number;
	    resource_id?: number;
	    type: string;
	    status: string;
	    workflow_run_id: number;
	    commit: string;
	    branch: string;
	    build_hash: string;
	    started_at: time.Time;
	    completed_at?: time.Time;
	    created_at: time.Time;
	    updated_at: time.Time;
	    service_name?: string;
	    resource_name?: string;
	
	    static createFrom(source: any = {}) {
	        return new ActionWithDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repository_id = source["repository_id"];
	        this.service_id = source["service_id"];
	        this.resource_id = source["resource_id"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.workflow_run_id = source["workflow_run_id"];
	        this.commit = source["commit"];
	        this.branch = source["branch"];
	        this.build_hash = source["build_hash"];
	        this.started_at = this.convertValues(source["started_at"], time.Time);
	        this.completed_at = this.convertValues(source["completed_at"], time.Time);
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	        this.service_name = source["service_name"];
	        this.resource_name = source["resource_name"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class KubernetesResource {
	    id: number;
	    repository_id: number;
	    name: string;
	    path: string;
	    resource_type: string;
	    namespace: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new KubernetesResource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repository_id = source["repository_id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.resource_type = source["resource_type"];
	        this.namespace = source["namespace"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Microservice {
	    id: number;
	    repository_id: number;
	    name: string;
	    path: string;
	    description: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Microservice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repository_id = source["repository_id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.description = source["description"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Project {
	    id: number;
	    name: string;
	    description: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Repository {
	    id: number;
	    name: string;
	    url: string;
	    type: string;
	    description: string;
	    service_name?: string;
	    service_location?: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	    last_sync_at?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Repository(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.service_name = source["service_name"];
	        this.service_location = source["service_location"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	        this.last_sync_at = this.convertValues(source["last_sync_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Task {
	    id: number;
	    project_id: number;
	    jira_ticket_id: string;
	    title: string;
	    description: string;
	    scheduled_date?: time.Time;
	    deadline?: time.Time;
	    status: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.jira_ticket_id = source["jira_ticket_id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.scheduled_date = this.convertValues(source["scheduled_date"], time.Time);
	        this.deadline = this.convertValues(source["deadline"], time.Time);
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TaskWithProject {
	    id: number;
	    project_id: number;
	    jira_ticket_id: string;
	    title: string;
	    description: string;
	    scheduled_date?: time.Time;
	    deadline?: time.Time;
	    status: string;
	    created_at: time.Time;
	    updated_at: time.Time;
	    project_name: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskWithProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.jira_ticket_id = source["jira_ticket_id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.scheduled_date = this.convertValues(source["scheduled_date"], time.Time);
	        this.deadline = this.convertValues(source["deadline"], time.Time);
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	        this.project_name = source["project_name"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

