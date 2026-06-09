export namespace main {
	
	export class SelectedLogView {
	    index: number;
	    timeText: string;
	    level: string;
	    tag: string;
	    message: string;
	    source: string;
	    raw: string;
	    display: string;
	
	    static createFrom(source: any = {}) {
	        return new SelectedLogView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.timeText = source["timeText"];
	        this.level = source["level"];
	        this.tag = source["tag"];
	        this.message = source["message"];
	        this.source = source["source"];
	        this.raw = source["raw"];
	        this.display = source["display"];
	    }
	}
	export class LogItemView {
	    index: number;
	    timeText: string;
	    level: string;
	    tag: string;
	    message: string;
	    source: string;
	    raw: string;
	    display: string;
	    isMatch: boolean;
	    isCurrent: boolean;
	    isSelected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LogItemView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.timeText = source["timeText"];
	        this.level = source["level"];
	        this.tag = source["tag"];
	        this.message = source["message"];
	        this.source = source["source"];
	        this.raw = source["raw"];
	        this.display = source["display"];
	        this.isMatch = source["isMatch"];
	        this.isCurrent = source["isCurrent"];
	        this.isSelected = source["isSelected"];
	    }
	}
	export class PauseView {
	    active: boolean;
	    bufferedCount: number;
	    droppedCount: number;
	
	    static createFrom(source: any = {}) {
	        return new PauseView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active = source["active"];
	        this.bufferedCount = source["bufferedCount"];
	        this.droppedCount = source["droppedCount"];
	    }
	}
	export class SearchView {
	    query: string;
	    matchIndexes: number[];
	    current: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.matchIndexes = source["matchIndexes"];
	        this.current = source["current"];
	    }
	}
	export class SavedFilterView {
	    id: string;
	    name: string;
	    packageName: string;
	    query: string;
	
	    static createFrom(source: any = {}) {
	        return new SavedFilterView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.packageName = source["packageName"];
	        this.query = source["query"];
	    }
	}
	export class FilterView {
	    draft: string;
	    applied: string;
	    error: string;
	    activeFilterId: string;
	    saved: SavedFilterView[];
	    history: string[];
	
	    static createFrom(source: any = {}) {
	        return new FilterView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.draft = source["draft"];
	        this.applied = source["applied"];
	        this.error = source["error"];
	        this.activeFilterId = source["activeFilterId"];
	        this.saved = this.convertValues(source["saved"], SavedFilterView);
	        this.history = source["history"];
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
	export class ProcessView {
	    pid: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	    }
	}
	export class PackageView {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new PackageView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	    }
	}
	export class DeviceView {
	    id: string;
	    model: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.model = source["model"];
	        this.status = source["status"];
	    }
	}
	export class AppState {
	    status: string;
	    adbStatus: string;
	    devices: DeviceView[];
	    selectedDevice: string;
	    packageScope: string;
	    packages: PackageView[];
	    selectedPackage: string;
	    processes: ProcessView[];
	    selectedProcess: string;
	    boundPids: number[];
	    totalLogs: number;
	    visibleCount: number;
	    filter: FilterView;
	    search: SearchView;
	    pause: PauseView;
	    selectedIndex: number;
	    logs: LogItemView[];
	    selectedLog?: SelectedLogView;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.adbStatus = source["adbStatus"];
	        this.devices = this.convertValues(source["devices"], DeviceView);
	        this.selectedDevice = source["selectedDevice"];
	        this.packageScope = source["packageScope"];
	        this.packages = this.convertValues(source["packages"], PackageView);
	        this.selectedPackage = source["selectedPackage"];
	        this.processes = this.convertValues(source["processes"], ProcessView);
	        this.selectedProcess = source["selectedProcess"];
	        this.boundPids = source["boundPids"];
	        this.totalLogs = source["totalLogs"];
	        this.visibleCount = source["visibleCount"];
	        this.filter = this.convertValues(source["filter"], FilterView);
	        this.search = this.convertValues(source["search"], SearchView);
	        this.pause = this.convertValues(source["pause"], PauseView);
	        this.selectedIndex = source["selectedIndex"];
	        this.logs = this.convertValues(source["logs"], LogItemView);
	        this.selectedLog = this.convertValues(source["selectedLog"], SelectedLogView);
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

