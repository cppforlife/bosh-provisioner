package packagescompiler

import (
	"fmt"

	boshcomp "github.com/cloudfoundry/bosh-agent/agent/compiler"
	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bpagentclient "github.com/cppforlife/bosh-provisioner/agent/client"
	bpeventlog "github.com/cppforlife/bosh-provisioner/eventlog"
	bpcpkgsrepo "github.com/cppforlife/bosh-provisioner/packagescompiler/compiledpackagesrepo"
	bppkgsrepo "github.com/cppforlife/bosh-provisioner/packagescompiler/packagesrepo"
	bprel "github.com/cppforlife/bosh-provisioner/release"
)

const concretePackagesCompilerLogTag = "ConcretePackagesCompiler"

type ConcretePackagesCompiler struct {
	agentClient          bpagentclient.Client
	packagesRepo         bppkgsrepo.PackagesRepository
	compiledPackagesRepo bpcpkgsrepo.CompiledPackagesRepository
	blobstore            boshblob.Blobstore

	eventLog bpeventlog.Log
	logger   boshlog.Logger
}

func NewConcretePackagesCompiler(
	agentClient bpagentclient.Client,
	packagesRepo bppkgsrepo.PackagesRepository,
	compiledPackagesRepo bpcpkgsrepo.CompiledPackagesRepository,
	blobstore boshblob.Blobstore,
	eventLog bpeventlog.Log,
	logger boshlog.Logger,
) ConcretePackagesCompiler {
	return ConcretePackagesCompiler{
		agentClient:          agentClient,
		packagesRepo:         packagesRepo,
		compiledPackagesRepo: compiledPackagesRepo,
		blobstore:            blobstore,

		eventLog: eventLog,
		logger:   logger,
	}
}

func (pc ConcretePackagesCompiler) ApplyPrecompiledPackages(release bprel.Release) error {
	for _, pkg := range release.CompiledPackages {
		blobID, fingerprint, err := pc.blobstore.Create(pkg.TarPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating compiled package blob %s", pkg.Name)
		}

		pkgRec := bpcpkgsrepo.CompiledPackageRecord{
			BlobID: blobID,
			SHA1:   fingerprint,
		}

		err = pc.compiledPackagesRepo.Save(*pkg, pkgRec)
		if err != nil {
			return bosherr.WrapErrorf(err, "Saving compiled package %s", pkg.Name)
		}
	}

	return nil
}

// Compile populates blobstore with compiled packages for a given release packages.
// All packages are compiled regardless if they will be later used or not.
// Currently Compile does not account for stemcell differences.
func (pc ConcretePackagesCompiler) Compile(release bprel.Release) error {
	packages := release.ResolvedPackageDependencies()

	releaseDesc := fmt.Sprintf("Compiling release %s/%s", release.Name, release.Version)

	stage := pc.eventLog.BeginStage(releaseDesc, len(packages))

	for _, pkg := range packages {
		pkgDesc := fmt.Sprintf("%s/%s", pkg.Name, pkg.Version)

		task := stage.BeginTask(fmt.Sprintf("Package %s", pkgDesc))

		_, found, err := pc.compiledPackagesRepo.Find(*pkg)
		if err != nil {
			return task.End(bosherr.WrapErrorf(err, "Finding compiled package %s", pkg.Name))
		} else if found {
			task.End(nil)
			continue
		}

		err = task.End(pc.compilePkg(*pkg))
		if err != nil {
			return err
		}
	}

	return nil
}

// FindCompiledPackage returns previously compiled package for a given template.
// If such compiled package is not found, error is returned.
func (pc ConcretePackagesCompiler) FindCompiledPackage(pkg bprel.Package) (CompiledPackageRecord, error) {
	var compiledPkgRec CompiledPackageRecord

	rec, found, err := pc.compiledPackagesRepo.Find(pkg)
	if err != nil {
		return compiledPkgRec, bosherr.WrapErrorf(err, "Finding compiled package %s", pkg.Name)
	} else if !found {
		return compiledPkgRec, bosherr.Errorf("Expected to find compiled package %s", pkg.Name)
	}

	compiledPkgRec.SHA1 = rec.SHA1
	compiledPkgRec.BlobID = rec.BlobID

	return compiledPkgRec, nil
}

// compilePackage populates blobstore with a compiled package for a
// given package. Assumes that dependencies of given package have
// already been compiled and are in the blobstore.
func (pc ConcretePackagesCompiler) compilePkg(pkg bprel.Package) error {
	pc.logger.Debug(concretePackagesCompilerLogTag,
		"Preparing to compile package %v", pkg)

	pkgRec, found, err := pc.packagesRepo.Find(pkg)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding package source blob %s", pkg.Name)
	}

	if !found {
		blobID, fingerprint, err := pc.blobstore.Create(pkg.TarPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating package source blob %s", pkg.Name)
		}

		pkgRec = bppkgsrepo.PackageRecord{
			BlobID: blobID,
			SHA1:   fingerprint,
		}

		err = pc.packagesRepo.Save(pkg, pkgRec)
		if err != nil {
			return bosherr.WrapErrorf(err, "Saving package record %s", pkg.Name)
		}
	}

	deps, err := pc.buildPkgDeps(pkg)
	if err != nil {
		return err
	}

	compiledPkgRes, err := pc.agentClient.CompilePackage(
		pkgRec.BlobID, // source tar
		pkgRec.SHA1,   // source tar
		pkg.Name,
		pkg.Version,
		deps,
	)
	if err != nil {
		return bosherr.WrapErrorf(err, "Compiling package %s", pkg.Name)
	}

	compiledPkgRec := bpcpkgsrepo.CompiledPackageRecord{
		BlobID: compiledPkgRes.BlobID,
		SHA1:   compiledPkgRes.SHA1,
	}

	err = pc.compiledPackagesRepo.Save(pkg, compiledPkgRec)
	if err != nil {
		return bosherr.WrapErrorf(err, "Saving compiled package %s", pkg.Name)
	}

	return nil
}

// buildPkgDeps prepares dependencies for agent's compile_package.
// Assumes that all package dependencies were already compiled.
func (pc ConcretePackagesCompiler) buildPkgDeps(pkg bprel.Package) (boshcomp.Dependencies, error) {
	deps := boshcomp.Dependencies{}

	for _, depPkg := range pkg.Dependencies {
		compiledPkgRec, found, err := pc.compiledPackagesRepo.Find(*depPkg)
		if err != nil {
			return deps, bosherr.WrapErrorf(err, "Finding compiled package %s", depPkg.Name)
		} else if !found {
			return deps, bosherr.Errorf("Expected to find compiled package %s", depPkg.Name)
		}

		deps[depPkg.Name] = boshcomp.Package{
			Name:        depPkg.Name,
			Version:     depPkg.Version,
			BlobstoreID: compiledPkgRec.BlobID, // compiled tar
			Sha1:        compiledPkgRec.SHA1,   // compiled tar
		}
	}

	return deps, nil
}
