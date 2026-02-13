export namespace main {
	
	export class AssetDTO {
	    id: number;
	    itemId: number;
	    type: string;
	    name: string;
	    filePath: string;
	    fileSize: number;
	    fileHash: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AssetDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.itemId = source["itemId"];
	        this.type = source["type"];
	        this.name = source["name"];
	        this.filePath = source["filePath"];
	        this.fileSize = source["fileSize"];
	        this.fileHash = source["fileHash"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class GossipInfoResponse {
	    instance_id: string;
	    instance_name: string;
	    last_sync?: time.Time;
	    item_count: number;
	
	    static createFrom(source: any = {}) {
	        return new GossipInfoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instance_id = source["instance_id"];
	        this.instance_name = source["instance_name"];
	        this.last_sync = this.convertValues(source["last_sync"], time.Time);
	        this.item_count = source["item_count"];
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
	export class ItemDTO {
	    id: number;
	    name: string;
	    category: string;
	    brand: string;
	    model: string;
	    serialNumber: string;
	    purchaseDate?: string;
	    photoPath: string;
	    notes: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.brand = source["brand"];
	        this.model = source["model"];
	        this.serialNumber = source["serialNumber"];
	        this.purchaseDate = source["purchaseDate"];
	        this.photoPath = source["photoPath"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ItemWithAssetsDTO {
	    item: ItemDTO;
	    assets: AssetDTO[];
	    health: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemWithAssetsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item = this.convertValues(source["item"], ItemDTO);
	        this.assets = this.convertValues(source["assets"], AssetDTO);
	        this.health = source["health"];
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
	export class PeerDTO {
	    id: string;
	    name: string;
	    address: string;
	    lastSeen: string;
	    lastSync: string;
	    isTrusted: boolean;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new PeerDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.lastSeen = source["lastSeen"];
	        this.lastSync = source["lastSync"];
	        this.isTrusted = source["isTrusted"];
	        this.status = source["status"];
	    }
	}
	export class SyncLogDTO {
	    id: number;
	    peerName: string;
	    timestamp: string;
	    itemsReceived: number;
	    itemsSent: number;
	    conflicts: number;
	    durationMs: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncLogDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.peerName = source["peerName"];
	        this.timestamp = source["timestamp"];
	        this.itemsReceived = source["itemsReceived"];
	        this.itemsSent = source["itemsSent"];
	        this.conflicts = source["conflicts"];
	        this.durationMs = source["durationMs"];
	        this.error = source["error"];
	    }
	}
	export class SyncResultDTO {
	    itemsReceived: number;
	    itemsSent: number;
	    conflicts: number;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemsReceived = source["itemsReceived"];
	        this.itemsSent = source["itemsSent"];
	        this.conflicts = source["conflicts"];
	        this.durationMs = source["durationMs"];
	    }
	}

}

export namespace models {
	
	export class SyncResult {
	    ItemsReceived: number;
	    ItemsSent: number;
	    Conflicts: number;
	    DurationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ItemsReceived = source["ItemsReceived"];
	        this.ItemsSent = source["ItemsSent"];
	        this.Conflicts = source["Conflicts"];
	        this.DurationMs = source["DurationMs"];
	    }
	}

}

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

