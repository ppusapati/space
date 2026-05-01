# eo_ml

Earth-observation deep-learning inference (EO-FR-020/021/022/024/025).

Provides:

* Sliding-window tiling utilities (`preprocess.tile_image`,
  `preprocess.untile_segmentation`).
* Channels-first / channels-last conversion and per-band normalisation.
* Non-Maximum Suppression for detection (`postprocess.nms`).
* Segmentation post-processing (`postprocess.argmax_segmentation`).
* High-level wrappers — `Detector`, `Segmenter`, `Classifier` — that take
  an ONNX model and apply the standard pre/post-processing.

Model weights are **not** bundled; deployments supply trained `.onnx`
files (e.g. YOLOv8, Faster R-CNN, U-Net, DeepLabV3+, Vision
Transformer).
