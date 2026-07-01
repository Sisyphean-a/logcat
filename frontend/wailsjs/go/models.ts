export namespace app {
	
	export class SavedFilterDraft {
	    ExistingID: string;
	    Name: string;
	    PackageName: string;
	    Query: string;
	
	    static createFrom(source: any = {}) {
	        return new SavedFilterDraft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ExistingID = source["ExistingID"];
	        this.Name = source["Name"];
	        this.PackageName = source["PackageName"];
	        this.Query = source["Query"];
	    }
	}

}

export namespace main {
	
	export class SelectedLogView {
	    sourceIndex: number;
	    timeText: string;
	    level: string;
	    tag: string;
	    message: string;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new SelectedLogView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceIndex = source["sourceIndex"];
	        this.timeText = source["timeText"];
	        this.level = source["level"];
	        this.tag = source["tag"];
	        this.message = source["message"];
	        this.source = source["source"];
	    }
	}
	export class LogItemView {
	    sourceIndex: number;
	    timeText: string;
	    level: string;
	    tag: string;
	    message: string;
	    isFocused: boolean;
	    isSelected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LogItemView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceIndex = source["sourceIndex"];
	        this.timeText = source["timeText"];
	        this.level = source["level"];
	        this.tag = source["tag"];
	        this.message = source["message"];
	        this.isFocused = source["isFocused"];
	        this.isSelected = source["isSelected"];
	    }
	}
	export class PauseView {
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PauseView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active = source["active"];
	    }
	}
	export class SearchView {
	    query: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
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
	    defaultFilterId: string;
	    saved: SavedFilterView[];
	
	    static createFrom(source: any = {}) {
	        return new FilterView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.draft = source["draft"];
	        this.applied = source["applied"];
	        this.error = source["error"];
	        this.activeFilterId = source["activeFilterId"];
	        this.defaultFilterId = source["defaultFilterId"];
	        this.saved = this.convertValues(source["saved"], SavedFilterView);
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
	    revision: number;
	    status: string;
	    adbStatus: string;
	    sessionActive: boolean;
	    devices: DeviceView[];
	    selectedDevice: string;
	    packageScope: string;
	    packages: PackageView[];
	    selectedPackage: string;
	    totalLogs: number;
	    visibleCount: number;
	    filter: FilterView;
	    search: SearchView;
	    pause: PauseView;
	    selectedCount: number;
	    logs: LogItemView[];
	    selectedLog?: SelectedLogView;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.revision = source["revision"];
	        this.status = source["status"];
	        this.adbStatus = source["adbStatus"];
	        this.sessionActive = source["sessionActive"];
	        this.devices = this.convertValues(source["devices"], DeviceView);
	        this.selectedDevice = source["selectedDevice"];
	        this.packageScope = source["packageScope"];
	        this.packages = this.convertValues(source["packages"], PackageView);
	        this.selectedPackage = source["selectedPackage"];
	        this.totalLogs = source["totalLogs"];
	        this.visibleCount = source["visibleCount"];
	        this.filter = this.convertValues(source["filter"], FilterView);
	        this.search = this.convertValues(source["search"], SearchView);
	        this.pause = this.convertValues(source["pause"], PauseView);
	        this.selectedCount = source["selectedCount"];
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
	
	
	
	export class LogSelectionRequest {
	    index: number;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new LogSelectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.mode = source["mode"];
	    }
	}
	
	
	
	
	
	export class SelectionPatch {
	    revision: number;
	    selectedCount: number;
	    focusedSourceIndex: number;
	    selectedSourceIndexes: number[];
	    selectedLog?: SelectedLogView;
	
	    static createFrom(source: any = {}) {
	        return new SelectionPatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.revision = source["revision"];
	        this.selectedCount = source["selectedCount"];
	        this.focusedSourceIndex = source["focusedSourceIndex"];
	        this.selectedSourceIndexes = source["selectedSourceIndexes"];
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
