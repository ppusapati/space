"""Earth-observation deep-learning inference."""
from .preprocess import (
    BandStats,
    normalise,
    tile_image,
    to_channels_first,
    to_channels_last,
    untile_segmentation,
)
from .postprocess import (
    Detection,
    argmax_segmentation,
    nms,
)
from .pipeline import Classifier, Detector, Segmenter

__all__ = [
    "BandStats",
    "normalise",
    "tile_image",
    "to_channels_first",
    "to_channels_last",
    "untile_segmentation",
    "Detection",
    "argmax_segmentation",
    "nms",
    "Classifier",
    "Detector",
    "Segmenter",
]
