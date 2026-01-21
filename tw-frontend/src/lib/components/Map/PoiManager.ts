import { ActionManager } from "@babylonjs/core/Actions/actionManager";
import { ExecuteCodeAction } from "@babylonjs/core/Actions/directActions";
import { Color3 } from "@babylonjs/core/Maths/math.color";
import { Vector3 } from "@babylonjs/core/Maths/math.vector";
import { Mesh } from "@babylonjs/core/Meshes/mesh";
import { MeshBuilder } from "@babylonjs/core/Meshes/meshBuilder";
import { Scene } from "@babylonjs/core/scene";
import { AdvancedDynamicTexture, Control, Rectangle, TextBlock } from "@babylonjs/gui";
import type { PointOfInterest, POIType } from "$lib/types/pois";
import { StandardMaterial } from "@babylonjs/core/Materials/standardMaterial";

/**
 * Manages Point of Interest markers on the globe.
 */
export class PoiManager {
    private scene: Scene;
    private rootNode: Mesh;
    private markers: Map<string, Mesh> = new Map();
    private labels: Map<string, Control> = new Map();
    private guiTexture: AdvancedDynamicTexture;
    private radius: number;
    private onPoiClick: ((poi: PointOfInterest) => void) | undefined;

    constructor(scene: Scene, radius: number, onPoiClick?: (poi: PointOfInterest) => void) {
        this.scene = scene;
        this.radius = radius;
        this.onPoiClick = onPoiClick;
        this.rootNode = new Mesh("poiRoot", scene);
        this.rootNode.rotation.y = Math.PI; // Align with texture mapping if needed (often needed for equirectangular)
        this.guiTexture = AdvancedDynamicTexture.CreateFullscreenUI("poiUI", true, scene);
    }

    public setRadius(radius: number) {
        this.radius = radius;
        // Update positions if radius changes (e.g. earth resize?)
        // Usually planet radius is constant, but if scaled:
        // Re-position all markers
    }

    public updatePOIs(pois: PointOfInterest[]) {
        // Simple diffing: remove all, add all (optimize later if needed)
        this.clear();

        pois.forEach(poi => {
            this.createMarker(poi);
        });
    }

    private clear() {
        this.markers.forEach(mesh => mesh.dispose());
        this.markers.clear();
        this.labels.forEach(control => control.dispose());
        this.labels.clear();
    }

    private createMarker(poi: PointOfInterest) {
        // Convert Lat/Lon to Vector3
        // Lat: -90 to 90, Lon: -180 to 180
        const latRad = poi.coordinates.lat * (Math.PI / 180);
        const lonRad = poi.coordinates.lon * (Math.PI / 180);

        // Standard spherical conversion (Y up)
        // x = r * cos(lat) * cos(lon)
        // y = r * sin(lat)
        // z = r * cos(lat) * sin(lon)
        // Note: Check coordinate system. Babylon usually Y up.
        // Texture mapping: U = (lon + 180)/360, V = (lat + 90)/180.
        // If texture is standard equirectangular starting at -180 lon (left),
        // we map 3D position to match.

        // Adjust for Babylon's coordinate system (often flipped Z or texture offset)
        // Assuming standard:
        const x = this.radius * Math.cos(latRad) * Math.cos(lonRad);
        const y = this.radius * Math.sin(latRad);
        const z = this.radius * Math.cos(latRad) * Math.sin(lonRad);

        // Create marker mesh
        const marker = MeshBuilder.CreateSphere("poi_" + poi.id, { diameter: this.radius * 0.015 }, this.scene);
        marker.position = new Vector3(x, y, z);
        marker.parent = this.rootNode;

        // Color based on type
        const mat = new StandardMaterial("poiMat_" + poi.id, this.scene);
        mat.emissiveColor = this.getColorForType(poi.type);
        mat.disableLighting = true; // Always visible
        marker.material = mat;

        // Interaction
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
        if (this.labels.has(poi.id)) return;

        const rect = new Rectangle();
        rect.width = "200px";
        rect.height = "80px";
        rect.cornerRadius = 10;
        rect.color = "white";
        rect.thickness = 1;
        rect.background = "rgba(0, 0, 0, 0.7)";

        const label = new TextBlock();
        label.text = `${poi.name || "Unknown"}\n${poi.type}\nElev: ${Math.round(poi.elevation)}m`;
        label.color = "white";
        label.fontSize = 14;
        rect.addControl(label);

        this.guiTexture.addControl(rect);
        rect.linkWithMesh(mesh);
        rect.linkOffsetY = -50;

        this.labels.set(poi.id, rect);
    }

    private hideLabel(id: string) {
        const label = this.labels.get(id);
        if (label) {
            label.dispose();
            this.labels.delete(id);
        }
    }

    public setRootRotation(y: number) {
        this.rootNode.rotation.y = y;
    }

    public dispose() {
        this.clear();
        this.rootNode.dispose();
        this.guiTexture.dispose();
    }
}
