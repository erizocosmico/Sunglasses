var gulp = require('gulp');

var coffee = require('gulp-coffee');
var concat = require('gulp-concat');
var imagemin = require('gulp-imagemin');
var sass = require('gulp-sass');
var react = require('gulp-react');
var ngmin = require('gulp-ngmin');
var shell = require('gulp-shell');

var paths = {
    scripts: ['js/**/*.coffee', '!vendor/**/*.coffee'],
    images: 'images/**/*',
    sass: 'sass/**/*.scss',
    templates: 'templates/**/*.jsx'
};

gulp.task('setup', function () {
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
        .pipe(gulp.dest('../public/js'));
});

// Copy all static images
gulp.task('images', function() {
    return gulp.src(paths.images)
        .pipe(imagemin({optimizationLevel: 5}))
        .pipe(gulp.dest('../public/images'));
});

// Compile scss files
// TODO reactivate sass
gulp.task('sass', function () {
    gulp.src(paths.sass)
        .pipe(sass({
            outputStyle: 'compressed',
            errLogToConsole: gulp.env.watch
        }))
        .pipe(gulp.dest('../public/css'));
});

// Compile react templates
gulp.task('react', function () {
    return gulp.src(paths.templates)
        .pipe(react())
        .pipe(gulp.dest('../public/templates'));
});

// Move dependencies
gulp.task('vendor', function () {
    return gulp.src('vendor/**/*')
        .pipe(gulp.dest('../public/vendor'));
});

// Rerun the task when a file changes
gulp.task('watch', function() {
    gulp.watch(paths.scripts, ['scripts']);
    gulp.watch(paths.images, ['images']);
    gulp.watch(paths.templates, ['react']);
    // TODO uncomment this as soon as https://github.com/hcatlin/libsass/issues/331 is solved
    //gulp.watch(paths.sass, ['sass']);
    gulp.watch('vendor/**/*', ['vendor']);
});

// The default task (called when you run `gulp` from cli)
gulp.task('default', ['setup', 'scripts', 'images', 'react', 'vendor', 'watch']);
