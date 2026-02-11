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

}

