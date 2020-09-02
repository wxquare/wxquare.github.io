##一、KCF（kernel Correlation Filter）

- 就是需要设计一个滤波模板，使得当它作用在跟踪目标上时，得到的响应最大，最大响应值的位置就是目标的位置
 
- 基于检测到的目标进行跟踪，首先在跟踪之前对目标进行检测，得到目标的位置，然后对目标进行学习，跟踪。

- 利用单通道的灰度特征改进为现在可以使用多通道的HOG特征
 

 
##二、循环矩阵 



## 3、opencv的静态编译和链接
=================可以成功编译和链接,运行=========================
    g++ kcf_danmu.cpp -Wl,-Bstatic -L/usr/local/lib -L/usr/local/share/OpenCV/3rdparty/lib -lopencv_stitching -lopencv_superres -lopencv_videostab -lopencv_aruco -lopencv_bgsegm -lopencv_bioinspired -lopencv_ccalib -lopencv_dnn_objdetect -lopencv_dpm -lopencv_face -lopencv_freetype -lopencv_fuzzy -lopencv_hdf -lopencv_hfs -lopencv_img_hash -lopencv_line_descriptor -lopencv_optflow -lopencv_reg -lopencv_rgbd -lopencv_saliency -lopencv_stereo -lopencv_structured_light -lopencv_phase_unwrapping -lopencv_surface_matching -lopencv_tracking -lopencv_datasets -lopencv_text -lopencv_dnn -lopencv_plot -lopencv_xfeatures2d -lopencv_shape -lopencv_video -lopencv_ml -lopencv_ximgproc -lopencv_xobjdetect -lopencv_objdetect -lopencv_calib3d -lopencv_features2d -lopencv_highgui -lopencv_videoio -lopencv_imgcodecs -lopencv_flann -lopencv_xphoto -lopencv_photo -lopencv_imgproc -lopencv_core -littnotify -llibprotobuf -lzlib -llibwebp -llibpng -llibtiff -llibjasper -lquirc -lippiw -lippicv -Wl,-Bdynamic -lgtk-x11-2.0 -lgdk-x11-2.0  -lpangocairo-1.0 -latk-1.0 -lcairo -lgdk_pixbuf-2.0 -lgio-2.0 -lpangoft2-1.0 -lpango-1.0 -lgobject-2.0 -lglib-2.0 -lfontconfig -lgthread-2.0 -lImath -lIlmImf -lIex -lHalf -lIlmThread -ldc1394 -lavcodec -lavformat -lavutil -lswscale -lavresample -lfreetype -lharfbuzz -lrt -lpthread -lz -ldl -lm

=================尝试减少依赖=========================
    g++ -O3 kcf_danmu.cpp -Wl,-Bstatic -L/usr/local/lib -L/usr/local/share/OpenCV/3rdparty/lib -lopencv_videostab -lopencv_ccalib -lopencv_dnn_objdetect -lopencv_dpm -lopencv_face -lopencv_freetype -lopencv_fuzzy -lopencv_hdf -lopencv_img_hash -lopencv_line_descriptor -lopencv_optflow -lopencv_reg -lopencv_rgbd -lopencv_saliency -lopencv_stereo -lopencv_structured_light -lopencv_phase_unwrapping -lopencv_surface_matching -lopencv_tracking -lopencv_datasets -lopencv_text -lopencv_dnn -lopencv_plot -lopencv_xfeatures2d -lopencv_shape -lopencv_video -lopencv_ml -lopencv_ximgproc -lopencv_xobjdetect -lopencv_objdetect -lopencv_calib3d -lopencv_features2d -lopencv_highgui -lopencv_videoio -lopencv_imgcodecs -lopencv_flann -lopencv_xphoto -lopencv_photo -lopencv_imgproc -lopencv_core -littnotify -lzlib -llibwebp -llibpng -llibtiff -llibjasper -lquirc -lippiw -lippicv -Wl,-Bdynamic -lgtk-x11-2.0 -lgdk-x11-2.0  -lpangocairLL -LRTo-1.0 -latk-1.0 -lcairo -lgdk_pixbuf-2.0 -lgio-2.0 -lpangoft2-1.0 -lpango-1.0 -lgobject-2.0 -lglib-2.0 -lfontconfig -lgthread-2.0 -lImath -lIlmImf -lIex -lHalf -lIlmThread -ldc1394 -lavcodec -lavformat -lavutil -lswscale -lharfbuzz -lrt -lpthread -lz -ldl -lm



## gpu 解码编译
版本：Video_Codec_SDK_8.1.24
sudo find /usr -name libavcodec.pc
sudo find /usr -name opencv.pc
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:/usr/local/lib64/pkgconfig
make
