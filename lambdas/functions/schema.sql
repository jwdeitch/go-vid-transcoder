CREATE TABLE videos (
  id                        CHAR(7) DEFAULT substring(md5(random() :: TEXT) FROM 0 FOR 6),
  video_title               TEXT,
  uploaded_at               TIMESTAMP WITH TIME ZONE,
  uploaded_by               CIDR,
  video_length_seconds      INT,
  thumb_count               INT,
  is_processing             BOOL,
  pre_transcode_size_bytes  BIGINT,
  post_transcode_size_bytes BIGINT,
  view_count                INT,
  notes                     TEXT,
  is_private                BOOL
);

select * from videos;