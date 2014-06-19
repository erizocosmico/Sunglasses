var gulp = require('gulp');

var coffee = require('gulp-coffee');
var concat = require('gulp-concat');
var imagemin = require('gulp-imagemin');
var sass = require('gulp-sass');
var ngmin = require('gulp-ngmin');
var shell = require('gulp-shell');
var jsonminify = require('gulp-jsonminify');
var uglify = require('gulp-uglify');
var minifyCSS = require('gulp-minify-css');

var paths = {
    scripts: ['js/**/*.coffee', '!vendor/**/*.coffee'],
    images: 'images/**/*',
    sass: 'sass/**/*.scss',
    templates: 'templates/**/*.jsx'
};

gulp.task('setup', function() {
    return gulp.src('./')
        .pipe(shell([
            'npm install',
            'bower install'
        ]));
})

gulp.task('scripts', function() {
    // Minify and copy all JavaScript (except vendor scripts)
    return gulp.src(paths.scripts)
        .pipe(coffee())
        .pipe(ngmin())
        .pipe(concat('app.min.js'))
        .pipe(uglify())
        .pipe(gulp.dest('../public/js'));
});

// Copy all static images
gulp.task('images', function() {
    return gulp.src(paths.images)
        .pipe(gulp.dest('../public/images'));
});

// Compile scss files
gulp.task('sass', function() {
    gulp.src(paths.sass)
        .pipe(sass({
            outputStyle: 'compressed',
            errLogToConsole: gulp.env.watch
        }))
        .pipe(gulp.dest('../public/css'));
});

// Copies ionicons font
gulp.task('icons', function() {
    gulp.src(['vendor/ionicons/fonts/*'])
        .pipe(gulp.dest('../public/fonts'));
});

// Generates a single css file with all vendor and site styles
gulp.task('prod-css', function() {
    gulp.src(['vendor/normalize.css/normalize.css',
        'vendor/semantic-ui/build/packaged/css/semantic.min.css',
        'vendor/ionicons/css/ionicons.min.css',
        'vendor/animate.css/animate.min.css',
        '../public/css/style.css'])
        .pipe(concat('style.min.css'))
        .pipe(minifyCSS({keepBreaks:false}))
        .pipe(gulp.dest('../public/css'));
});

// Generates a vendor.min.js file with all vendor dependencies
gulp.task('prod-js', function() {
    gulp.src(['vendor/angular/angular.min.js',
        'vendor/angular-route/angular-route.min.js',
        'vendor/angular-cookies/angular-cookies.min.js',
        'vendor/angular-translate/angular-translate.min.js',
        'vendor/angular-translate-storage-cookie/angular-translate-storage-cookie.min.js',
        'vendor/angular-translate-storage-local/angular-translate-storage-local.min.js',
        'vendor/angular-translate-loader-static-files/angular-translate-loader-static-files.min.js',
        'vendor/js-md5/js/md5.min.js',
        'vendor/jquery/dist/jquery.min.js',
        'vendor/semantic-ui/build/packaged/javascript/semantic.min.js'])
        .pipe(concat('vendor.min.js'))
        .pipe(gulp.dest('../public/js'));
});

// HTML templates
gulp.task('tpls', function() {
    return gulp.src('templates/**/*.html')
        .pipe(gulp.dest('../public/templates'));
});

// Copy index file
gulp.task('index', function() {
    return gulp.src('app.html')
        .pipe(gulp.dest('../public'))
})

// Move dependencies
gulp.task('vendor', function() {
    return gulp.src('vendor/**/*')
        .pipe(gulp.dest('../public/vendor'));
});

// Move language files
gulp.task('lang', function() {
    return gulp.src('lang/*.json')
        .pipe(jsonminify())
        .pipe(gulp.dest('../public/lang'));
});

// Rerun the task when a file changes
gulp.task('watch', function() {
    gulp.watch(paths.scripts, ['scripts']);
    gulp.watch(paths.images, ['images']);
    gulp.watch(paths.templates, ['react']);
    gulp.watch(paths.sass, ['sass']);
    gulp.watch('app.html', ['index']);
    gulp.watch('lang/*.json', ['lang']);
    gulp.watch('templates/**/*.html', ['tpls']);
    gulp.watch('../public/css/style.css', ['prod-css']);
});

// The default task (called when you run `gulp` from cli)
gulp.task('default', ['index', 'scripts', 'sass', 'lang', 'images', 'tpls', 'icons', 'prod-css', 'prod-js']);
