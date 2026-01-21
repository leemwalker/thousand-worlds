import { ActionManager } from "@babylonjs/core/Actions/actionManager";
import { ExecuteCodeAction } from "@babylonjs/core/Actions/directActions";
import { Color3 } from "@babylonjs/core/Maths/math.color";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Mesh } from "@babylonjs/core/Meshes/mesh";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { Scene } from "@babylonjs/core/scene";
// import { AdvancedDynamicTexture } from "@babylonjs/gui/2D/advancedDynamicTexture.js";
// import { Control } from "@babylonjs/gui/2D/controls/control.js";
// import { Rectangle } from "@babylonjs/gui/2D/controls/rectangle.js";
// import { TextBlock } from "@babylonjs/gui/2D/controls/textBlock.js";
import type { PointOfInterest, POIType } from "$lib/types/pois";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";

/**
 * Manages Point of Interest markers on the globe.
 */
export class PoiManager {
    private scene: Scene;
    private rootNode: Mesh;
    private markers: Map<string, Mesh> = new Map();
    // private labels: Map<string, Control> = new Map();
    // private guiTexture: AdvancedDynamicTexture; // GUI disabled to fix build
    private radius: number;
    private onPoiClick: ((poi: PointOfInterest) => void) | undefined;

    constructor(scene: Scene, radius: number, onPoiClick?: (poi: PointOfInterest) => void) {
        this.scene = scene;
        this.radius = radius;
        this.onPoiClick = onPoiClick;
        this.rootNode = new Mesh("poiRoot", scene);
        this.rootNode.rotation.y = Math.PI;
        // this.guiTexture = AdvancedDynamicTexture.CreateFullscreenUI("poiUI", true, scene);
    }

    public setRadius(radius: number) {
        this.radius = radius;
    }

    public updatePOIs(pois: PointOfInterest[]) {
        this.clear();
        pois.forEach(poi => {
            this.createMarker(poi);
        });
    }

    private clear() {
        this.markers.forEach(mesh => mesh.dispose());
        this.markers.clear();
        // this.labels.forEach(control => {
        //     // control.dispose() 
        // });
        // this.labels.clear();
    }

    private createMarker(poi: PointOfInterest) {
        // ... (coordinates code unchanged) ...
        const latRad = poi.coordinates.lat * (Math.PI / 180);
        const lonRad = poi.coordinates.lon * (Math.PI / 180);

        const x = this.radius * Math.cos(latRad) * Math.cos(lonRad);
        const y = this.radius * Math.sin(latRad);
        const z = this.radius * Math.cos(latRad) * Math.sin(lonRad);

        // Create marker mesh
        const marker = MeshBuilder.CreateSphere("poi_" + poi.id, { diameter: this.radius * 0.015 }, this.scene);
        marker.position = new Vector3(x, y, z);
        marker.parent = this.rootNode;

        const mat = new StandardMaterial("poiMat_" + poi.id, this.scene);
        mat.emissiveColor = this.getColorForType(poi.type);
        mat.disableLighting = true;
        marker.material = mat;

        marker.actionManager = new ActionManager(this.scene);
        marker.actionManager.registerAction(new ExecuteCodeAction(
            ActionManager.OnPointerOverTrigger,
            () => {
                this.showLabel(poi, marker);
            }
        ));
        marker.actionManager.registerAction(new ExecuteCodeAction(
            ActionManager.OnPickTrigger,
            () => {
                console.log("[PoiManager] Clicked POI:", poi.name);
                if (this.onPoiClick) {
                    this.onPoiClick(poi);
                }
            }
        ));
        marker.actionManager.registerAction(new ExecuteCodeAction(
            ActionManager.OnPointerOutTrigger,
            () => {
                this.hideLabel(poi.id);
            }
        ));

        this.markers.set(poi.id, marker);
    }

    private getColorForType(type: POIType): Color3 {
        switch (type) {
            case "mountain_peak": return Color3.White();
            case "volcano": return Color3.Red();
            case "deep_ocean": return Color3.Blue();
            default: return Color3.Yellow();
        }
    }

    private showLabel(poi: PointOfInterest, mesh: Mesh) {
        // GUI Disabled
        // if (this.labels.has(poi.id)) return;
        // ... implementation removed ...
    }

    private hideLabel(id: string) {
        // GUI Disabled
        // const label = this.labels.get(id);
        // if (label) {
        //     label.dispose();
        //     this.labels.delete(id);
        // }
    }

    public setRootRotation(y: number) {
        this.rootNode.rotation.y = y;
    }

    public dispose() {
        this.clear();
        this.rootNode.dispose();
        // this.guiTexture.dispose();
    }
}
